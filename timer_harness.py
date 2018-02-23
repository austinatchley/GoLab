#!/usr/bin/env python3

import sys
import csv
import subprocess
import numpy as np
import matplotlib.pyplot as plt

from io import StringIO

TESTS = 2

if len(sys.argv) < 5:
    print(
        "./harness [hash-workers] [data-workers] [comparison-workers] [input]"
    )
    sys.exit(0)

args = "./tree.out -hash-workers " + sys.argv[1] +      \
    "  -data-workers=" + sys.argv[2] +                  \
    " -comp-workers=" + sys.argv[3]  +                  \
    " -input=" + sys.argv[4]

h = sys.argv[1]
d = sys.argv[2]
c = sys.argv[3]
i = sys.argv[4]

times = []
times_spin = []


def do_test(i, h, d, c):
    arg_list = args.split()

    print(" ".join(str(val) for val in arg_list))

    ret_val = 0
    try:
        ret_val = run(arg_list)
    except:
        print("Error. Skipping test case")
        ret_val = -1
    return ret_val


def run(arg_list):
    output = subprocess.check_output(arg_list)
    result = output.decode('utf-8')

    f = StringIO(result)
    reader = csv.reader(f, delimiter=',')
    data = []
    for row in reader:
        data.append(' '.join(element.rstrip() for element in row))

    print('Data:\t', data)

    return 1


def control_test():
    arg_list = args.split()

    output = subprocess.check_output(arg_list)
    result = output.decode('utf-8')

    f = StringIO(result)
    reader = csv.reader(f, delimiter=',')
    data = []
    for row in reader:
        data.append(' '.join(element.rstrip() for element in row))

    print("data:\t", data)
    return 1


print("Control")
control_test()
control = control_test()
print("")

both = 0.0
both_spin = 0.0

tests_completed = 0
tests_completed_spin = 0

for i in range(TESTS):
    print("\nIteration ", i)
    val = do_test(i, h, d, c)
    if val != -1:
        both += val
        tests_completed += 1
for i in range(TESTS):
    print("\nIteration ", i)
    val = do_test(i, h, d, c)
    if val != -1:
        both_spin += val
        tests_completed_spin += 1

average = both / tests_completed
average_spin = both_spin / tests_completed_spin
print(average, average_spin)

print('\nSpeedups:')
print(average)
print(average_spin)




"""
for core in range(1, cores + 1):
    do_test(-1, core, False)

    both = 0.0
    both_spin = 0.0
    tests_completed = 0
    tests_completed_spin = 0

    for i in range(TESTS):
        print("\nIteration ", i, "with ", core, "cores. Mutex")
        val = do_test(i, core, False)
        if val != -1:
            both += val
            tests_completed += 1
    for i in range(TESTS):
        print("\nIteration ", i, "with ", core, "cores. Spinlock")
        val = do_test(i, core, True)
        if val != -1:
            both_spin += val
            tests_completed_spin += 1

    average = both / tests_completed
    average_spin = both_spin / tests_completed_spin

    times.insert(core, control / average)
    times_spin.insert(core, control / average_spin)

print('\nSpeedups:')
for time in times:
    print(time)

color = '#1f10e0'
color_spin = '#ff0011'

plt.plot(
    list(map(int, range(1,
                        int(cores) + 1))),
    times,
    c=color,
    alpha=0.8,
    marker='o')
plt.plot(
    list(map(int, range(1,
                        int(cores) + 1))),
    times_spin,
    c=color_spin,
    alpha=0.8,
    marker='o')

plt.legend(['Mutex', 'Spinlock'])

plt.xlabel('Number of Cores')
plt.ylabel('Speedup')
#plt.show()
plt.savefig("graph" + version + ".pdf", bbox_inches='tight', format='pdf')
"""
