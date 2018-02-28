#!/usr/bin/env python3

import sys
import csv
import subprocess
import numpy as np
import matplotlib.pyplot as plt

from io import StringIO

TESTS = 2

if len(sys.argv) < 4:
    print(
        "./harness [hash-workers] [data-workers] [comparison-workers]"
    )
    sys.exit(0)

args = "./step2.out -hash-workers=" +      \
    "  -data-workers=" + sys.argv[2] +                  \
    " -comp-workers=" + sys.argv[3]  +                  \
    " -input="

h = int(sys.argv[1])
d = int(sys.argv[2])
c = int(sys.argv[3])

hash_times = []
hash_insert_times = []
total_times = []

hash_times_l = []
hash_insert_times_l = []
total_times_l = []

def do_test(i, h, d, c, extra):
    arg_list = args.split()
    arg_list[1] += str(h)
    arg_list[4] += str(i)
    arg_list.append(str(extra))

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
    reader = csv.reader(f, delimiter='\n')
    data = []
    for row in reader:
        data.append(' '.join(element.rstrip() for element in row))

    print('Data:\t', data)
    data = [float(x) for x in data]

    return data

def control_test(i):
    arg_list = args.split()
    arg_list[1] += "1"
    arg_list[4] += str(i)

    output = subprocess.check_output(arg_list)
    result = output.decode('utf-8')

    f = StringIO(result)
    reader = csv.reader(f, delimiter='\n')
    data = []
    for row in reader:
        data.append(' '.join(element.rstrip() for element in row))

    print("data:\t", data)
    data = [float(x) for x in data]

    return data

def plot_bars(barGroups, barNames, groupNames, colors, ylabel="", title="", width=0.8):
    """Plot a grouped bar chart
    barGroups  - list of groups, where each group is a list of bar heights
    barNames   - list containing the name of each bar within any group
    groupNames - list containing the name of each group
    colors     - list containing the color for each bar within a group
    ylabel     - label for the y-axis
    title      - title
    """
    fig, ax = plt.subplots()
    offset = lambda items, off: [x + off for x in items]

    maxlen = max(len(group) for group in barGroups)
    xvals = range(len(barGroups))
    
    for i, bars in enumerate(zip(*barGroups)):
        print(bars)
        plt.bar(offset(xvals, i * width/maxlen), bars, width/maxlen, color=colors[i])

    ax.set_ylabel(ylabel)
    ax.set_title(title)
    ax.set_xticks(offset(xvals, width / 2))
    ax.set_xticklabels(groupNames)

    # Shrink current axis by 20%
    box = ax.get_position()
    ax.set_position([box.x0, box.y0, box.width * 0.8, box.height])

    # Put a legend to the right of the current axis
    ax.legend(barNames, loc="upper left", bbox_to_anchor=(1, 1))



print("Control")
control_test("sample/coarse.txt")
control = control_test("sample/coarse.txt")
control_test("sample/fine.txt")
control_fine = control_test("sample/fine.txt")
print("control coarse: ", control)
print("control fine: ", control_fine)

nums = []

workers = 1
it = 0
while workers <= h:
    nums.append(workers)
    hash_time = 0.0
    hash_insert_time = 0.0
    total_time = 0.0

    tests_completed = 0

    for i in range(TESTS):
        print("\nIteration ", i)
        val = do_test("sample/coarse.txt", workers, d, c, "")
        if val != -1:
            hash_time += val[0]
            hash_insert_time += val[1]
            total_time += val[2]

            tests_completed += 1

    average_hash = hash_time / tests_completed
    average_hash_insert = hash_insert_time / tests_completed
    average_total = total_time / tests_completed

    hash_times.insert(it, control[0] / average_hash)
    hash_insert_times.insert(it, control[1] / average_hash_insert)
    total_times.insert(it, control[2] / average_total)

    # Do the fine test now
    hash_time = 0.0
    hash_insert_time = 0.0
    total_time = 0.0

    tests_completed = 0

    for i in range(TESTS):
        print("\nIteration ", i)
        val = do_test("sample/fine.txt", workers, d, c, "-l")
        if val != -1:
            hash_time += val[0]
            hash_insert_time += val[1]
            total_time += val[2]

            tests_completed += 1

    average_hash = hash_time / tests_completed
    average_hash_insert = hash_insert_time / tests_completed
    average_total = total_time / tests_completed

    hash_times_l.insert(it, control_fine[0] / average_hash)
    hash_insert_times_l.insert(it, control_fine[1] / average_hash_insert)
    total_times_l.insert(it, control_fine[2] / average_total)

    workers = workers * 2
    it+=1

print('\nSpeedups:')

print("Coarse:\nHashing")
for i in range(0, len(hash_times)):
    print(i, hash_times[i])

print("Hashing+Insertion")
for i in range(0, len(hash_insert_times)):
    print(i, hash_insert_times[i])

print("Total")
for i in range(0, len(total_times)):
    print(i, total_times[i])

print("Fine:\nHashing")
for i in range(0, len(hash_times_l)):
    print(i, hash_times_l[i])

print("Hashing+Insertion")
for i in range(0, len(hash_insert_times_l)):
    print(i, hash_insert_times_l[i])

print("Total")
for i in range(0, len(total_times_l)):
    print(i, total_times_l[i])


"""
plt.bar(nums, hash_times) 
plt.show()

plt.bar(nums, hash_insert_times)
plt.show()

plt.bar(nums, total_times)
plt.show()
"""

control_data = [control[0], control[1], control[2] - control[1]]
control_fine_data = [control_fine[0], control_fine[1], control_fine[2] - control_fine[1]]

plot_bars([control_data, control_fine_data], ["Hash Time", "Hash and Insert Time", "Comparison Time"], ["coarse", "fine"], ["#5caec4", "#c0cccf", "#2f3638"], "Time (nanoseconds)", "Sequential Runtime")
plt.savefig("sequential.pdf", bbox_inches='tight', format='pdf')
#plt.show()
plt.close()

cm = plt.get_cmap('plasma')
colors = [cm(i / len(nums)) for i in range(0, len(nums))]

plot_bars([hash_times, hash_times_l], [str(num) + " Workers" for num in nums], ["coarse", "fine"], colors, "Speedup", "Step 2: Hashing Speedup")
plt.savefig("hash.pdf", bbox_inches='tight', format='pdf')
#plt.show()
plt.close()

plot_bars([hash_insert_times, hash_insert_times_l], [str(num) + " Workers" for num in nums], ["coarse", "fine"], colors, "Speedup", "Step 2: Hashing + Insertion Speedup")
plt.savefig("hashinsert.pdf", bbox_inches='tight', format='pdf')
#plt.show()
plt.close()

plot_bars([total_times, total_times_l], [str(num) + " Workers" for num in nums], ["coarse", "fine"], colors, "Speedup", "Step 2: Total Speedup")
plt.savefig("total.pdf", bbox_inches='tight', format='pdf')
#plt.show()
plt.close()


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
