import numpy as np
import matplotlib.pyplot as plt

# Sigmoid function
def sigmoid(x, k=1):
    x = x / 200
    return 1 / (1 + np.exp(-k * (x - 0.5)))

x = np.arange(1, 201)
y = sigmoid(x, k=10)  # k controls the steepness

plt.plot(x, y)
plt.title("Sigmoid Curve")
plt.xlabel("x")
plt.ylabel("sigmoid(x)")
plt.grid(True)
plt.show()