package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*
* Structs
 */

type Tree struct {
	Left  *Tree
	Value int
	Right *Tree
}

type Pair struct {
	i int
	j int
}

/*
* Struct Methods
 */

func (tree *Tree) AddNode(val int) (e error) {
	e = nil

	switch {
	case val < tree.Value:
		if tree.Left != nil {
			tree.Left.AddNode(val)
		} else {
			left := Tree{nil, val, nil}
			tree.Left = &left
		}
	case val > tree.Value:
		if tree.Right != nil {
			tree.Right.AddNode(val)
		} else {
			right := Tree{nil, val, nil}
			tree.Right = &right
		}
	default:
		e = errors.New("given value is equal to the value of the current node")
	}
	return
}

func (tree *Tree) Length() int {
	left, right := 1, 1
	if tree.Left != nil {
		left = tree.Left.Length()
	}
	if tree.Right != nil {
		right = tree.Right.Length()
	}

	if left > right {
		return left
	}
	return right
}

func (tree *Tree) Hash() uint32 {
	c := make(chan int)
	// go
	go Walk(tree, c, nil)

	var hash uint32 = 0
	var prime uint32 = 4222234741

	// Unpack values from the channel c while it is OK
	val, ok := <-c
	for ; ok; val, ok = <-c {
		val2 := uint32(val) + 2
		hash = (hash*val2 + val2) % prime
	}
	return hash
}

/*
* Main
 */

func main() {
	var hWorkers, dWorkers, cWorkers int
	flag.IntVar(&hWorkers, "hash-workers", 1, "number of workers on hashing")
	flag.IntVar(&dWorkers, "data-workers", 1, "number of workers on data")
	flag.IntVar(&cWorkers, "comp-workers", 1, "number of workers on comparison")

	var input string
	flag.StringVar(&input, "input", "", "path to input file")

	var lockVar, pMatrix bool
	flag.BoolVar(&lockVar, "l", false, "lock on hashMap write")
	flag.BoolVar(&pMatrix, "p", false, "print the resulting equivalency matrix")

	flag.Parse()

	_, _, _, _ = hWorkers, dWorkers, cWorkers, input

  // This is the main container for all the trees
	trees := make([]Tree, 0)

  // Reads and constructs all trees from the input
	readInput(&trees, input)

	// hashMap is a map from hash to slice of tree indices
	// hashes is the slice of hashes by index mirroring that of trees
	hashMap := make(map[uint32][]int, len(trees))
	hashes := make([]uint32, len(trees))

	// matrix is the adjacency matrix, initialized to false
	matrix := make([][]bool, len(trees))

	// Construct adjacency matrix columns
	for i := range matrix {
		matrix[i] = make([]bool, len(trees))
	}

	var lock sync.Mutex

	/*
		  // Spawn a goroutine for each tree
			for i, tree := range trees {
		    go func(tree *Tree, i int){
		      hashes[i] = tree.Hash()
		    }(&tree, i)
		  }

		  for i, hash := range hashes {
		    hashMap[hash] = append(hashMap[hash], i)
		  }
	*/
	pairChan := make(chan struct {
		uint32
		int
	}, len(trees))

  // If we have more workers than trees, just assign a worker to each tree
	if hWorkers > len(trees) {
		hWorkers = len(trees)
	}

	finishedHashMap := make(chan int, hWorkers)
	finishedHashTimer := make(chan int, hWorkers)

	// Create dummy hash slice to isolate and time hashing algorithm
	dummyHashes := make([]uint32, len(trees))
	startHashing := time.Now()
	if !lockVar {
		hashChan := make(chan *[]uint32, hWorkers)

		for i := 0; i < hWorkers; i++ {
			curTrees := trees[computeBounds(len(trees), i, hWorkers):computeBounds(len(trees), i+1, hWorkers)]
			//fmt.Println(curTrees)
			go computeHashesParallel(&curTrees, computeBounds(len(trees), i, hWorkers), hashChan)
		}

		for i := 0; i < hWorkers; i++ {
			hash := <-hashChan
			dummyHashes = append(dummyHashes, *hash...)
		}
	} else {
		for i := 0; i < hWorkers; i++ {
			curTrees := trees[computeBounds(len(trees), i, hWorkers):computeBounds(len(trees), i+1, hWorkers)]
			go computeHashes(&curTrees, computeBounds(len(trees), i, hWorkers), &dummyHashes, finishedHashTimer)
		}
		for i := 0; i < hWorkers; i++ {
			<-finishedHashTimer
		}
	}
	endHashing := time.Now()
	hashingTime := endHashing.Sub(startHashing).Nanoseconds()
	fmt.Println(hashingTime)

  // make sure we don't have any extraneous data that will affect timing
	runtime.GC()
	// Actually compute the hashing and insert in the map
	// Start timing for total and hash+insert
	beginTime := time.Now()
	beginHashingPlusInsert := time.Now()

	// If we aren't locking, start receiving hashes to insert in the map before they are sent
	if !lockVar {
		go insertHashesSingle(pairChan, finishedHashMap, &hashMap, len(trees))
	}

	if hWorkers == 1 {
		if lockVar {
			computeHashesLock(&trees, 0, &hashes, &hashMap, &lock, finishedHashMap)
			<-finishedHashMap
		} else {
			hashChan := make(chan *[]uint32, len(trees))
			go computeHashesSingle(&trees, 0, hashChan, pairChan)
			hashes = *(<-hashChan)
		}
	} else {
		if lockVar {
			for i := 0; i < hWorkers; i++ {
        // Partition each worker's piece of the trees slice
				curTrees := trees[computeBounds(len(trees), i, hWorkers):computeBounds(len(trees), i+1, hWorkers)]
				go computeHashesLock(&curTrees, computeBounds(len(trees), i, hWorkers), &hashes, &hashMap, &lock, finishedHashMap)
			}

      // Make sure that each worker has finished before continuing
			for i := 0; i < hWorkers; i++ {
				<-finishedHashMap
			}
		} else {
			hashChan := make(chan *[]uint32, hWorkers)

			for i := 0; i < hWorkers; i++ {
				curTrees := trees[computeBounds(len(trees), i, hWorkers):computeBounds(len(trees), i+1, hWorkers)]

        // Computes the hashes for its partition and adds to a private slice
        // Also sends insertHashesSingle() the data it has been waiting for
				go computeHashesSingle(&curTrees, computeBounds(len(trees), i, hWorkers), hashChan, pairChan)
			}

      // Merge the private slices
			for i := 0; i < hWorkers; i++ {
				hash := <-hashChan
				hashes = append(hashes, *hash...)
			}
			<-finishedHashMap
		}
	}
	endHashingPlusInsert := time.Now()
	hashingPlusInsertTime := endHashingPlusInsert.Sub(beginHashingPlusInsert).Nanoseconds()
	fmt.Println(hashingPlusInsertTime)

	var wg sync.WaitGroup

	// Fill diagonal with true
	for i := range matrix {
		matrix[i][i] = true
	}

	if cWorkers == 1 {            // Do the sequential algorithm
		for _, list := range hashMap {
			for i := range list {
        // We can start from i + 1 because Same(i,i) is always true
				for j := i + 1; j < len(list); j++ {
					li := list[i]
					lj := list[j]
					// Compare the supposed equivalent trees
					result := SameTraverse(&trees[li], &trees[lj])

					// Mirror result to cut down on computation
					matrix[li][lj] = result
					matrix[lj][li] = result
				}
			}
		}
	} else if cWorkers == -1 {    // Unlimited goroutines approach
		for _, list := range hashMap {
			//fmt.Println(list)
			for i := range list {
				for j := i + 1; j < len(list); j++ {
          // For each goroutine created, add 1 to the wg
					wg.Add(1)
					go func(li, lj int) {
            // Tell the WaitGroup we are done after returning
						defer wg.Done()
						result := SameTraverse(&trees[li], &trees[lj])

						// Mirror result to cut down on computation
						matrix[li][lj] = result
						matrix[lj][li] = result
					}(list[i], list[j])
				}
			}
		}
	} else {                    // Thread Pool Approach
    // Pair is a pair of int indices in trees
		treeChan := make(chan Pair, len(hashMap))

		wg.Add(cWorkers)
		for i := 0; i < cWorkers; i++ {
      // Tell cWorkers to start looking for data in treeChan
			go parallelComparison(treeChan, &trees, &matrix, &wg)
		}

    // For every pair that needs to be checked, put it in treeChan
    // Each of the cWorkers will fight for the data in the channel
		for _, list := range hashMap {
			for i := range list {
				for j := i + 1; j < len(list); j++ {
					treeChan <- Pair{list[i], list[j]}
				}
			}
		}
    // This acts as a stopper at the end of the list
    // One for each cWorker to read
		for i := 0; i < cWorkers; i++ {
			treeChan <- Pair{-1, -1}
		}
	}
  // Wait until we are completely done inserting
	wg.Wait()

	endTime := time.Now()

	if pMatrix {
		printMatrix(&matrix)
	}

	diff := endTime.Sub(beginTime).Nanoseconds()
	fmt.Println(diff)
}

/*
* Functions
 */

// Recursively walk over the tree and put the vals in ch
func _walk(t *Tree, ch chan int) {
	if t != nil {
		_walk(t.Left, ch)
		ch <- t.Value
		_walk(t.Right, ch)
	}
}

// See _walk
// If something is pulled out of kill, preempts walk and returns
func _walk_preempt(t *Tree, ch, kill chan int) {
	if t != nil {
		_walk_preempt(t.Left, ch, kill)
		select {
		case <-kill:
			return
		default:
			ch <- t.Value
		}
		_walk_preempt(t.Right, ch, kill)
	}
}

// Walk walks the tree t sending all values
// from the tree to the channel ch.
// If a kill chan is provided, we will preempt
// whenever a value appears in kill
func Walk(t *Tree, ch, kill chan int) {
	if kill == nil {
		_walk(t, ch)
	} else {
		_walk_preempt(t, ch, kill)
	}
	close(ch)
}

// Determines whether the trees
// t1 and t2 contain the same values.
func SameTraverse(t1, t2 *Tree) bool {
	c1 := make(chan int)
	c2 := make(chan int)
	kill := make(chan int, 2)

	go Walk(t1, c1, kill)
	go Walk(t2, c2, kill)

	for {
		v1, ok1 := <-c1
		v2, ok2 := <-c2

		if v1 != v2 || ok1 != ok2 {
      // Before returning, kill the walking goroutine
			kill <- 1
			kill <- 2
			return false
		}

		if !ok1 {
			break
		}
	}
  // See :379
	kill <- 1
	kill <- 2
	return true
}

func createTree(data *[]int) (tree *Tree, e error) {
	if len(*data) == 0 {
		e = errors.New("length of data is 0")
		return
	}

	treeVal := Tree{nil, (*data)[0], nil}
	tree = &treeVal
	for i := 1; i < len(*data); i++ {
		err := tree.AddNode((*data)[i])
		check(err)
	}
	return
}

func computeHashes(trees *[]Tree, offset int, hashes *[]uint32, finished chan int) {
	for i, elem := range *trees {
		hash := elem.Hash()
		(*hashes)[i+offset] = hash
	}
	finished <- 1
}

func computeHashesParallel(trees *[]Tree, offset int, hashChan chan *[]uint32) {
	hashes := make([]uint32, len(*trees))
	for i, elem := range *trees {
		hash := elem.Hash()
		hashes[i] = hash
	}
	if hashChan != nil {
		hashChan <- &hashes
	}
}

func computeHashesLock(trees *[]Tree, offset int, hashes *[]uint32, hashMap *map[uint32][]int, lock *sync.Mutex, finished chan int) {
	for i, elem := range *trees {
		hash := elem.Hash()
		(*hashes)[i+offset] = hash
	}

	for i := offset; i < offset+len(*trees); i++ {
		lock.Lock()
		(*hashMap)[(*hashes)[i]] = append((*hashMap)[(*hashes)[i]], i)
		lock.Unlock()
	}
	finished <- 1
}

// pairChan sends data to insertHashesSingle()
func computeHashesSingle(trees *[]Tree, offset int, hashChan chan *[]uint32, pairChan chan struct {
	uint32
	int
}) {
	hashes := make([]uint32, len(*trees))
	for i, elem := range *trees {
		hash := elem.Hash()
		hashes[i] = hash
		pairChan <- struct {
			uint32
			int
		}{hash, i + offset}
	}
	if hashChan != nil {
		hashChan <- &hashes
	}
}

func insertHashesSingle(pairChan chan struct {
	uint32
	int
}, finished chan int, hashMap *map[uint32][]int, length int) {
	for i := 0; i < length; i++ {
		pair := <-pairChan
		(*hashMap)[pair.uint32] = append((*hashMap)[pair.uint32], pair.int)
	}
	finished <- 1
}

func parallelComparison(treeChan chan Pair, trees *[]Tree, matrix *[][]bool, wg *sync.WaitGroup) {
	for {
		pair, ok := <-treeChan
		if !ok {
			break
		}
		i := pair.i
		j := pair.j

    // This is the end case. Tell wg we are done
		if i == -1 && j == -1 {
			break
		}

		result := SameTraverse(&(*trees)[i], &(*trees)[j])

    // Mirror result across diagonal
		(*matrix)[i][j] = result
		(*matrix)[j][i] = result
	}
	wg.Done()
}

/*
* Utility Functions
 */

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func readInput(trees *[]Tree, input string) {
	f, err := os.Open(input)
	check(err)
	reader := bufio.NewReader(f)

	index := 0
	rawData, err := reader.ReadString('\n')
	for ; err == nil; rawData, err = reader.ReadString('\n') {
		check(err)

		strData := strings.Fields(rawData)
		data := make([]int, len(strData))
		for i, elem := range strData {
			data[i], err = strconv.Atoi(elem)
			check(err)
		}

		tree, err := createTree(&data)
		check(err)
		*trees = append(*trees, *tree)
		index++

		// fmt.Println(data, len(data))
	}
}

func printMatrix(matrix *[][]bool) {
	for _, elem := range *matrix {
		fmt.Println(elem)
	}
}

func computeBounds(lenTrees int, i int, hWorkers int) int {
	return (lenTrees * i) / hWorkers
}
