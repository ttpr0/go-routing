import matplotlib.pyplot as plt
import numpy as np

Y_AXIS = "Zeit [s]"
X_AXIS = ""
LEGEND = []

x_data = []
y_data = []

with open("./results_1800.csv", "r") as file:
    lines = file.readlines()
    tokens = lines[0].strip().split(";")
    X_AXIS = tokens[0]
    LEGEND = tokens[1:]

    for line in lines[1:]:
        tokens = line.strip().split(";")
        x_data.append(int(tokens[0]))
        y_data.append([int(t) for t in tokens[1:]])

x_data = np.array(x_data)
y_data = np.array(y_data)


plt.style.use('ggplot')

fig = plt.figure()
ax = fig.add_subplot(111)
ax.set_xlabel(X_AXIS)
ax.set_ylabel(Y_AXIS)
ax.grid(visible=True)

for i in range(y_data.shape[1]):
    ax.plot(x_data, y_data[:, i] / 1000,
            label=LEGEND[i], linestyle='-', marker='o')

ax.legend()

plt.show()
