#!/usr/bin/env python3

import sys
import re
import numpy as np
from typing import List, Set, Dict, Tuple, Iterator, Union, Optional, Callable, Any, IO
from sklearn.preprocessing import LabelEncoder, OneHotEncoder
import tensorflow as tf

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
	return np.array(out)

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
		one_hot_encoded = one_hot_encoder.fit_transform(integer_encoded)
		input_features.append(one_hot_encoded.toarray())

	input_features = np.stack(input_features)
	return integer_encoded, one_hot_encoded, input_features

def label_encode(sequences1: List[str], sequences2: List[str]) -> Any:
	integer_encoded1, one_hot_encoded1, input_features1 = label_encode_seq(sequences1)
	integer_encoded2, one_hot_encoded2, input_features2 = label_encode_seq(sequences2)

	one_hot_combo = np.stack([input_features1, input_features2], axis = 2)
	return one_hot_combo

def get_all_data(seqpath1: str, seqpath2: str, pairpath: str, col: int) -> Tuple[Any, Any]:
	seq1, seq2, pairs = get_seqs_and_pair(seqpath1, seqpath2, pairpath, col)
	onehot = label_encode(seq1, seq2)
	return (onehot, pairs)

def build_model() -> Any:
	model = tf.keras.models.Sequential([
	  tf.keras.layers.Flatten(input_shape=(25, 2, 4)),
	  # tf.keras.layers.Flatten(input_shape=(28, 28)),
	  tf.keras.layers.Dense(128, activation='relu'),
	  tf.keras.layers.Dropout(0.2),
	  tf.keras.layers.Dense(10)
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

def measure_loss(loss_fn, y_train, predictions):
	return loss_fn(y_train[:1], predictions).numpy()

def compile(model, loss_fn):
	model.compile(optimizer='adam',
		      loss=loss_fn,
		      metrics=['accuracy'])

def fit(model, x_train, y_train):
	model.fit(x_train, y_train, epochs=5)

def evaluate_after_fit(model, x_test, y_test):
	print(model.evaluate(x_test,  y_test, verbose=2))

def make_prob_model(model):
	probability_model = tf.keras.Sequential([
	  model,
	  tf.keras.layers.Softmax()
	])
	return probability_model

def show_first_5_prob_model_probs(probability_model, x_test):
	print(probability_model(x_test[:5]))

def split_train_vs_test(x: Any, y: Any) -> Tuple[Any, Any, Any, Any]:
	half = len(x) // 2
	return (x[:half], x[half:], y[:half], y[half:])

def main():
	seqpath1: str = sys.argv[1]
	seqpath2: str = sys.argv[2]
	pairpath: str = sys.argv[3]
	paircol: int = int(sys.argv[4])

	print("got args")

	x, y = get_all_data(seqpath1, seqpath2, pairpath, paircol)
	x_train, x_test, y_train, y_test = split_train_vs_test(x, y)

	print("got data")

	model = build_model()
	loss_fn = get_loss_fn()

	print("build model and loss")

	compile(model, loss_fn)
	fit(model, x_train, y_train)
	print("model:", model)
	print("x_test:", x_test)
	print("y_test:", y_test)
	evaluate_after_fit(model, x_test, y_test)

	print("fit and tested")

	prob_model = make_prob_model(model)
	print("made prob")
	show_first_5_prob_model_probs(prob_model, x_test)
	print("printed")

if __name__ == "__main__":
	main()
