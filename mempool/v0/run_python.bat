@echo off
call D:\Anaconda3\Scripts\activate.bat D:\Anaconda3
call conda activate myenv
cd D:\GitHubProject\tendermint_work\mempool\v0
python D:\GitHubProject\tendermint_work\mempool\v0\scripts.py D:\GitHubProject\tendermint_work\mempool\v0\neural_net_regression.pkl

@REM cd C:\HarddiskD\program\python\test\test
@REM call chdir