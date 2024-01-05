#!/usr/bin/env python3

import sys
import re
import numpy as np
from typing import List, Set, Dict, Tuple, Iterator, Union, Optional, Callable, Any, IO
from sklearn.preprocessing import LabelEncoder, OneHotEncoder
import tensorflow as tf
import random

class FaEntry:
	def __init__(self, header: str, seq: str):
		self.seq: str = seq
		self.header: str = header

def get_data(r: IO) -> List[FaEntry]:
	header: str = ""
	seq: str = ""
	out: List[FaEntry] = []

	hre: Any = re.compile("""^>""")
	emptyre: Any = re.compile("""^$""")

	for line in r:
		line = line.rstrip("\n")
		if emptyre.search(line):
			continue
		if hre.search(line):
			if len(header) > 0 and len(seq) > 0:
				out.append(FaEntry(header, seq))
			header = line[1:]
			seq = ""
			continue
		seq = seq + line
	if len(header) > 0 and len(seq) > 0:
		out.append(FaEntry(header, seq))
	return out

def fa_to_seqs(fa: List[FaEntry]) -> List[str]:
	return [x.seq for x in fa]

def get_pair(path: str, col: int) -> Any:
	out: List[float] = []
	with open(path, "r") as r:
		for line in r:
			sl: List[str] = line.rstrip("\n").split("\t")
			if len(sl) <= col:
				raise Exception("len(sl) <= col")
			out.append(float(sl[col]))
	return out

def arrayify_pairs(pairs: List[float]) -> Any:
	return np.array(pairs)

def get_seqs_and_pair(fapath1: str, fapath2: str, pairpath: str, paircol: int) -> Tuple[List[str], List[str], Any]:
	with open(fapath1, "r") as r:
		seqs1 = fa_to_seqs(get_data(r))
	with open(fapath2, "r") as r:
		seqs2 = fa_to_seqs(get_data(r))
	pairs = get_pair(pairpath, paircol)
	return (seqs1, seqs2, pairs)

def label_encode_seq(sequences: List[str]) -> Tuple[Any, Any, List[Any]]:
	# The LabelEncoder encodes a sequence of bases as a sequence of integers.
	integer_encoder = LabelEncoder()	

	# The OneHotEncoder converts an array of integers to a sparse matrix where 
	# each row corresponds to one possible value of each feature.
	one_hot_encoder = OneHotEncoder(categories='auto')	 
	input_features = []

	for sequence in sequences:
		integer_encoded = integer_encoder.fit_transform(list(sequence))
		integer_encoded = np.array(integer_encoded).reshape(-1, 1)
		print("integer_encoded:", integer_encoded)
		one_hot_encoded = one_hot_encoder.fit_transform(integer_encoded)
		input_features.append(one_hot_encoded.toarray())

	input_features = np.stack(input_features)
	return integer_encoded, one_hot_encoded, input_features

def enc(base: str) -> int:
	return ["a", "t", "g", "c", "n"].index(base)

def label_encode_seq_v2(sequences: List[str]) -> Tuple[Any, Any, List[Any]]:
	# The OneHotEncoder converts an array of integers to a sparse matrix where 
	# each row corresponds to one possible value of each feature.
	one_hot_encoder = OneHotEncoder(categories='auto')	 
	input_features = []

	for sequence in sequences:
		# df['col1_num'] = df['col1'].apply(lambda x: ['first', 'second', 'third', 'fourth'].index(x))

		integer_encoded = [enc(x) for x in list(sequence.lower())]
		# integer_encoded = integer_encoder.fit_transform(list(sequence.lower()))
		integer_encoded2: Any = np.array(integer_encoded).reshape(-1, 1)
		one_hot_encoded = one_hot_encoder.fit_transform(integer_encoded2)
		input_features.append(one_hot_encoded.toarray())

	print(input_features)
	input_features = np.stack(input_features)
	return integer_encoded, one_hot_encoded, input_features

def label_encode_seq_v3(sequences: List[str]) -> List[Any]:
	# The OneHotEncoder converts an array of integers to a sparse matrix where 
	# each row corresponds to one possible value of each feature.
	one_hot_encoder = OneHotEncoder(categories=[[0], [1], [2], [3], [4]])
	input_features = []

	for sequence in sequences:
		integer_encoded = [enc(x) for x in list(sequence.lower())]
		# integer_encoded = integer_encoder.fit_transform(list(sequence.lower()))
		integer_encoded2: Any = np.array(integer_encoded).reshape(-1, 1)

		print("integer_encoded2:", integer_encoded2)
		one_hot_encoded = one_hot_encoder.fit_transform(integer_encoded2)
		input_features.append(one_hot_encoded.toarray())

	print(input_features)
	input_features = np.stack(input_features)
	return input_features

def label_encode_seq_v4(seqs1: List[str], seqs2: List[str]) -> List[Any]:
	one_hot_encoder = OneHotEncoder(categories = ['a', 't', 'g', 'c', 'n'])
	unencoded = []
	for seq1, seq2 in zip(seqs1, seqs2):
		unencoded.append([
			one_hot_encoder.fit_transform([ [x] for x in list(seq1.lower())]),
			one_hot_encoder.fit_transform([ [x] for x in list(seq2.lower())])
		])
	return np.stack(unencoded)

def encodeseqs(o: Any, seq1: str, seq2: str) -> Any:
	a1: Any = np.array(list(seq1.lower())).reshape(-1, 1)
	a2: Any = np.array(list(seq2.lower())).reshape(-1, 1)

	e1: Any = np.array(o.fit_transform(a1).todense())
	e2: Any = np.array(o.fit_transform(a2).todense())

	out: Any = np.stack([e1, e2], axis = 1)
	return out

def label_encode_seq_v5(seqs1: List[str], seqs2: List[str]) -> List[Any]:
	o = OneHotEncoder(categories = [['a', 't', 'g', 'c', 'n']])
	unstacked = [encodeseqs(o, x, y) for x, y in zip(seqs1, seqs2)]
	out = np.stack(unstacked, axis = 0)
	return out

def stack_encoded_v5(unstacked: List[Any]) -> Any:
	return np.stack(unstacked, axis = 0)

def label_encode(sequences1: List[str], sequences2: List[str]) -> List[Any]:
	return label_encode_seq_v5(sequences1, sequences2)


def get_all_data(seqpath1: str, seqpath2: str, pairpath: str, col: int) -> Tuple[List[Any], Any]:
	seq1, seq2, pairs = get_seqs_and_pair(seqpath1, seqpath2, pairpath, col)
	onehot = label_encode(seq1, seq2)
	return (onehot, pairs)

def build_model_unedited(length: int) -> Any:
	model = tf.keras.models.Sequential([
		tf.keras.layers.Flatten(input_shape=(length, 2, 5)),
		# tf.keras.layers.Flatten(input_shape=(28, 28)),
		tf.keras.layers.Dense(128, activation='relu'),
		tf.keras.layers.Dropout(0.2),

		# tf.keras.layers.Dense(10)
		tf.keras.layers.Dense(1, activation = 'linear')
	])
	return model

def build_model_conv(length: int) -> Any:
	model = tf.keras.models.Sequential([
		tf.keras.layers.Conv1D(32, kernel_size = 2, strides = 3, input_shape = (length, 2, 5)),
		tf.keras.layers.Flatten(),

		tf.keras.layers.Dense(128, activation='relu'),
		tf.keras.layers.Dense(32, activation='softplus'),

		tf.keras.layers.Dropout(0.2),

		# tf.keras.layers.Dense(10)
		tf.keras.layers.Dense(1, activation = 'linear')
	])
	return model

def build_model_lstm(length: int) -> Any:
	model = tf.keras.models.Sequential([
		tf.keras.layers.LSTM(64, input_shape = (length, 2, 5)),
		tf.keras.layers.Flatten(),

		tf.keras.layers.Dense(128, activation='relu'),
		tf.keras.layers.Dense(32, activation='softplus'),

		tf.keras.layers.Dropout(0.2),

		# tf.keras.layers.Dense(10)
		tf.keras.layers.Dense(1, activation = 'linear')
	])
	return model

def build_model(length: int) -> Any:
	model = tf.keras.models.Sequential([
		tf.keras.layers.Flatten(input_shape=(length, 2, 5)),
		# tf.keras.layers.Flatten(input_shape=(28, 28)),
		tf.keras.layers.Dense(128, activation='relu'),
		tf.keras.layers.Dense(32, activation='softplus'),
		tf.keras.layers.Dropout(0.2),

		# tf.keras.layers.Dense(10)
		tf.keras.layers.Dense(1, activation = 'linear')
	])
	return model

def predict(model, x_train):
	predictions = model(x_train[:1]).numpy()
	return predictions

def predictions_to_probabilities(predictions):
	return tf.nn.softmax(predictions).numpy()

def get_loss_fn():
	loss_fn = tf.keras.losses.SparseCategoricalCrossentropy(from_logits=True)
	return loss_fn

def get_loss_fn_mse():
	return "mse"

def measure_loss(loss_fn, y_train, predictions):
	return loss_fn(y_train[:1], predictions).numpy()

def compile_unedited(model, loss_fn):
	model.compile(optimizer='adam',
			loss=loss_fn,
			metrics=['accuracy'])

def compile(model, loss_fn):
	model.compile(optimizer='adam',
			loss=loss_fn)

def fit(model, x_train, y_train):
	model.fit(x_train, y_train, epochs=5)

def fitbig(model, x_train, y_train):
	model.fit(x_train, y_train, epochs=5, verbose = 0)

def evaluate_after_fit(model, x_test, y_test):
	print(model.evaluate(x_test,	y_test, verbose=2))

def make_prob_model(model):
	probability_model = tf.keras.Sequential([
		model,
		tf.keras.layers.Softmax()
	])
	return probability_model

def show_first_5_prob_model_probs(probability_model, x_test):
	print(probability_model(x_test[:5]))

def shuf(x: List[Any], y: List[float]) -> Tuple[List[Any], List[Any]]:
	# z = []
	# for xarg, yarg in zip(x, y):
	# 	z.append((xarg, yarg))
	z = list(zip(x, y))
	random.shuffle(z)

	x2: List[Any] = []
	y2: List[float] = []
	for val in z:
		x2.append(val[0])
		y2.append(val[1])
	return (x2, y2)

def split_train_vs_test(x: Any, y: Any) -> Tuple[Any, Any, Any, Any]:
	half = len(x) // 2
	return (x[:half], x[half:], y[:half], y[half:])

def main() -> None:
	seqpath1: str = sys.argv[1]
	seqpath2: str = sys.argv[2]
	pairpath: str = sys.argv[3]
	paircol: int = int(sys.argv[4])
	length: int = int(sys.argv[5])

	x, y = get_all_data(seqpath1, seqpath2, pairpath, paircol)
	x, y = shuf(x, y)
	x2 = stack_encoded_v5(x)
	y2 = arrayify_pairs(y)
	x_train, x_test, y_train, y_test = split_train_vs_test(x2, y2)

	model = build_model(length)
	loss_fn = get_loss_fn_mse()
	# loss_fn = get_loss_fn()

	compile(model, loss_fn)
	print("x_train:", x_train)
	print("y_train:", y_train)
	# fit(model, x_train, y_train)
	fitbig(model, x_train, y_train)
	print("model:", model)
	print("x_test:", x_test)
	print("y_test:", y_test)
	evaluate_after_fit(model, x_test, y_test)

	prob_model = make_prob_model(model)
	show_first_5_prob_model_probs(prob_model, x_test)

	y_predict = model.predict(x_test)
	y_predict_list = [x[0] for x in list(y_predict)]
	for i, j in zip(list(y_test), y_predict_list):
		print(i, j)

if __name__ == "__main__":
	main()
