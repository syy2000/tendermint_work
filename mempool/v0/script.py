import sys
import pickle
import numpy as np
from collections.abc import Iterable
import os
from keras.models import Sequential
from keras.layers import Dense

os.environ['TF_ENABLE_ONEDNN_OPTS'] = '0'
if __name__ == "__main__":
    
    np.set_printoptions(suppress=True, precision=2)
    data_list = []
    while True:
        line = sys.stdin.readline()
        if not line:
            break
        words = line.strip().split(" ")
        numpy_array = np.array([float(value) for value in words])
        data_list.append(numpy_array)
    data_array = np.array(data_list)
    #data = sys.stdin.readline().strip()  # 读取一行输入
    #values = data.split(' ')
    #numpy_array = np.array(np.array([float(value) for value in values]))
    #numpy_array = np.array([5000,0.1,0.2])
    #numpy_array = numpy_array.reshape(1, 3)
    #print(data_array)
    with open(sys.argv[1], 'rb') as f:
        model = pickle.load(f)
    #model = tf.keras.models.load_model(sys.argv[1])
    predictions = model.predict(data_array)
    print(predictions)
    np.savetxt("weight.txt", predictions, fmt="%.2f")
    #filename = 'numpy_data.txt'
    #np.save(filename, predictions, fmt='%d', delimiter=' ')