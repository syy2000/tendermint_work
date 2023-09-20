package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	cfg "github.com/tendermint/tendermint/config"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

const (
	exampleAddress          string = "192.168.128.129"
	exampleIPAddressAndPort string = exampleAddress + ":26657"
)

var (
	p2pIPAddresses, rpcIPAddresses                                  string
	allowEmptyBlock, allowRecheck, allowBroadcast                   bool
	p2pRate                                                         int64
	timeoutPropose, timeoutPrevote, timeoutPrecommit, timeoutCommit string
	mempoolSize                                                     int
)

func init() {
	InitLogFilesCmd.Flags().StringVar(&p2pIPAddresses, "p2pIP", "127.0.0.1:26656", "example:192.168.128.129:26656,192.168.128.130:26656,192.168.128.131:26656,192.168.128.132:26656")
	InitLogFilesCmd.Flags().StringVar(&rpcIPAddresses, "rpcIP", "127.0.0.1:26657", "example:192.168.128.129:26657,192.168.128.130:26657,192.168.128.131:26657,192.168.128.132:26657")
	InitLogFilesCmd.Flags().StringVar(&hostnamePrefix, "hostname-prefix", "node",
		"hostname prefix (\"node\" results in persistent peers list ID0@node0:26656, ID1@node1:26656, ...)")
	InitLogFilesCmd.Flags().StringVar(&hostnameSuffix, "hostname-suffix", "",
		"hostname suffix ("+
			"\".xyz.com\""+
			" results in persistent peers list ID0@node0.xyz.com:26656, ID1@node1.xyz.com:26656, ...)")
	InitLogFilesCmd.Flags().StringVar(&outputDir, "o", "./mytestnet",
		"directory to store initialization data for the testnet")
	InitLogFilesCmd.Flags().Int64Var(&initialHeight, "initial-height", 0,
		"initial height of the first block")
	InitLogFilesCmd.Flags().IntVar(&nValidators, "v", -1,
		"number of validators to initialize the testnet with")
	// worth modified
	InitLogFilesCmd.Flags().BoolVar(&allowEmptyBlock, "allow-empty", false, "if allow empty block")
	InitLogFilesCmd.Flags().Int64Var(&p2pRate, "p2p-rate", -1, "maximum p2p rate")
	InitLogFilesCmd.Flags().StringVar(&timeoutPropose, "timeout-propose", "3s", "timeout-propose")
	InitLogFilesCmd.Flags().StringVar(&timeoutPrevote, "timeout-prevote", "1s", "timeout-prevote")
	InitLogFilesCmd.Flags().StringVar(&timeoutPrecommit, "timeout-precommit", "1s", "timeout-precommit")
	InitLogFilesCmd.Flags().StringVar(&timeoutCommit, "timeout-commit", "1s", "timeout-commit")
	InitLogFilesCmd.Flags().BoolVar(&allowBroadcast, "allow-broadcast", true, "allow mempool broadcast or not")
	InitLogFilesCmd.Flags().BoolVar(&allowRecheck, "allow-recheck", false, "allow mempool recheck")
	InitLogFilesCmd.Flags().IntVar(&mempoolSize, "mempool-size", 10000, "mempool size")

}

var InitLogFilesCmd = &cobra.Command{
	Use:   "initConfig",
	Short: "Initialize config with p2p address and rpc address",
	Long: `initConfig command will initialize config files and genisis files for nodes.
	An example is tendermint initConfig --p2pIP 192.168.128.129:26656,192.168.128.130:26656,192.168.128.131:26656,192.168.128.132:26656 --rpcIP 192.168.128.129:26657,192.168.128.130:26657,192.168.128.131:26657,192.168.128.132:26657
	`,
	RunE: initLogFiles,
}

func initLogFiles(cmd *cobra.Command, args []string) error {
	config := cfg.DefaultConfig()
	p2ps, err := GetIP(p2pIPAddresses)
	if err != nil {
		return err
	}
	rpcs, err := GetIP(rpcIPAddresses)
	if err != nil {
		return err
	}
	if len(p2ps) != len(rpcs) {
		return fmt.Errorf("p2p and rpc must have the same number of ip addresses")
	}
	if nValidators < 0 {
		nValidators = len(p2ps)
		nNonValidators = 0
	} else if nValidators > len(p2ps) {
		return fmt.Errorf("nValidators %d is more than the number of ips : %d", nValidators, len(p2ps))
	} else {
		nNonValidators = len(p2ps) - nValidators
	}
	genVals := make([]types.GenesisValidator, nValidators)

	// validators
	for i := 0; i < nValidators; i++ {
		// example : node0
		nodeDirName := fmt.Sprintf("%s%d", nodeDirPrefix, i)
		nodeDir := filepath.Join(outputDir, nodeDirName)
		config.SetRoot(nodeDir)

		// mkdir
		err := os.MkdirAll(filepath.Join(nodeDir, "config"), nodeDirPerm)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}
		err = os.MkdirAll(filepath.Join(nodeDir, "data"), nodeDirPerm)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		if err := initFilesWithConfig(config); err != nil {
			return err
		}

		pvKeyFile := filepath.Join(nodeDir, config.BaseConfig.PrivValidatorKey)
		pvStateFile := filepath.Join(nodeDir, config.BaseConfig.PrivValidatorState)
		pv := privval.LoadFilePV(pvKeyFile, pvStateFile)

		pubKey, err := pv.GetPubKey()
		if err != nil {
			return fmt.Errorf("can't get pubkey: %w", err)
		}
		genVals[i] = types.GenesisValidator{
			Address: pubKey.Address(),
			PubKey:  pubKey,
			Power:   1,
			Name:    nodeDirName,
		}
	}

	// non-Validators
	for i := 0; i < nNonValidators; i++ {
		nodeDir := filepath.Join(outputDir, fmt.Sprintf("%s%d", nodeDirPrefix, i+nValidators))
		config.SetRoot(nodeDir)

		err := os.MkdirAll(filepath.Join(nodeDir, "config"), nodeDirPerm)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		err = os.MkdirAll(filepath.Join(nodeDir, "data"), nodeDirPerm)
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}

		if err := initFilesWithConfig(config); err != nil {
			return err
		}
	}

	// Generate genesis doc from generated validators
	genDoc := &types.GenesisDoc{
		ChainID:         "chain-" + tmrand.Str(6),
		ConsensusParams: types.DefaultConsensusParams(),
		GenesisTime:     tmtime.Now(),
		InitialHeight:   initialHeight,
		Validators:      genVals,
	}
	genDoc.ConsensusParams.Block.MaxBytes = 1024 * 1024

	// Write genesis file.
	for i := 0; i < nValidators+nNonValidators; i++ {
		nodeDir := filepath.Join(outputDir, fmt.Sprintf("%s%d", nodeDirPrefix, i))
		if err := genDoc.SaveAs(filepath.Join(nodeDir, config.BaseConfig.Genesis)); err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}
	}

	// Gather persistent peer addresses.
	var persistentPeers string
	genPersistentPeer := func() (string, error) {
		persistentPeers := make([]string, nValidators+nNonValidators)
		for i := 0; i < nValidators+nNonValidators; i++ {
			nodeDir := filepath.Join(outputDir, fmt.Sprintf("%s%d", nodeDirPrefix, i))
			config.SetRoot(nodeDir)
			nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
			if err != nil {
				return "", err
			}
			persistentPeers[i] = p2p.IDAddressString(nodeKey.ID(), p2ps[i])
		}
		return strings.Join(persistentPeers, ","), nil
	}
	if populatePersistentPeers {
		persistentPeers, err = genPersistentPeer()
		if err != nil {
			_ = os.RemoveAll(outputDir)
			return err
		}
	}

	// Overwrite default config.
	for i := 0; i < nValidators+nNonValidators; i++ {
		nodeDir := filepath.Join(outputDir, fmt.Sprintf("%s%d", nodeDirPrefix, i))
		config.SetRoot(nodeDir)
		config.P2P.AddrBookStrict = false
		config.P2P.AllowDuplicateIP = true
		if populatePersistentPeers {
			config.P2P.PersistentPeers = persistentPeers
		}
		// IP Ports
		// ip is validated,so no error here
		_, p2pport, _ := GetIPAndPort(p2ps[i])
		_, rpcport, _ := GetIPAndPort(rpcs[i])
		config.P2P.ListenAddress = "tcp://0.0.0.0:" + p2pport
		config.RPC.ListenAddress = "tcp://0.0.0.0:" + rpcport

		// user defined
		if p2pRate > 0 {
			config.P2P.RecvRate, config.P2P.SendRate = p2pRate, p2pRate
		}
		config.Consensus.CreateEmptyBlocks = allowEmptyBlock
		config.Mempool.Recheck, config.Mempool.Broadcast = allowRecheck, allowBroadcast
		config.Mempool.Size = mempoolSize
		config.Mempool.CacheSize = 50 * mempoolSize
		if len(timeoutCommit) > 0 {
			if tmp, err := time.ParseDuration(timeoutCommit); err == nil {
				config.Consensus.TimeoutCommit = tmp
			}
		}
		if len(timeoutCommit) > 0 {
			if tmp, err := time.ParseDuration(timeoutPrecommit); err == nil {
				config.Consensus.TimeoutPrecommit = tmp
			}
		}
		if len(timeoutPrevote) > 0 {
			if tmp, err := time.ParseDuration(timeoutPrevote); err == nil {
				config.Consensus.TimeoutPrevote = tmp
			}
		}
		if len(timeoutPropose) > 0 {
			if tmp, err := time.ParseDuration(timeoutPropose); err == nil {
				config.Consensus.TimeoutPropose = tmp
			}
		}

		config.Moniker = moniker(i)

		cfg.WriteConfigFile(filepath.Join(nodeDir, "config", "config.toml"), config)
	}

	fmt.Printf("Successfully initialized %v node directories\n", nValidators+nNonValidators)
	return nil
}

// ============== tools ========================================

func VerifyIPAddress(addr string) error {
	u := strings.Split(addr, ".")
	if len(u) != 4 {
		return fmt.Errorf("want ip address like %s, get %s", exampleAddress, addr)
	}
	for _, numStr := range u {
		if n, err := strconv.Atoi(numStr); err != nil {
			return fmt.Errorf("want ip address like %s, get %s", exampleAddress, addr)
		} else if n < 0 || n > 255 {
			return fmt.Errorf("ip address must be in 0-255, get %s", addr)
		}
	}
	return nil
}
func VerifyIPPort(port string) error {
	if n, err := strconv.Atoi(port); err != nil {
		return fmt.Errorf("want an integer port like 26657, get %s", port)
	} else if n < 0 || n > 65535 {
		return fmt.Errorf("ip port must be in 0-65535, get %s", port)
	}
	return nil
}
func VerifyIP(addr string) error {
	u := strings.Split(addr, ":")
	if len(u) != 2 {
		return fmt.Errorf("want ip address like %s, get %s", exampleIPAddressAndPort, addr)
	}
	if err := VerifyIPAddress(u[0]); err != nil {
		return err
	}
	if err := VerifyIPPort(u[1]); err != nil {
		return err
	}
	return nil
}
func GetIP(ips string) ([]string, error) {
	u := strings.Split(ips, ",")
	if len(u) == 0 {
		return nil, fmt.Errorf("an example of 4 nodes : 192.168.128.129:26657,192.168.128.130:26657,192.168.128.131:26657,192.168.128.132:26657")
	}
	for _, str := range u {
		if err := VerifyIP(str); err != nil {
			return nil, err
		}
	}
	return u, nil
}
func GetIPAndPort(ip string) (string, string, error) {
	if err := VerifyIP(ip); err != nil {
		return "", "", err
	}
	u := strings.Split(ip, ":")
	return u[0], u[1], nil

}
