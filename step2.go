package main

import (
  "runtime"
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
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
	go Walk(tree, c)

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
	var hWorkers int
	flag.IntVar(&hWorkers, "hash-workers", 1, "number of workers on hashing")

	var dWorkers int
	flag.IntVar(&dWorkers, "data-workers", 1, "number of workers on data")

	var cWorkers int
	flag.IntVar(&cWorkers, "comp-workers", 1, "number of workers on comparison")

	var input string
	flag.StringVar(&input, "input", "", "path to input file")

	var lockVar bool
	flag.BoolVar(&lockVar, "l", false, "lock on hashMap write")

	var pMatrix bool
	flag.BoolVar(&pMatrix, "p", false, "print the resulting equivalency matrix")

	flag.Parse()

	_, _, _, _ = hWorkers, dWorkers, cWorkers, input

	//fmt.Println("Number of Hash Workers: ", hWorkers)
	//fmt.Println("Number of Data Workers: ", *dWorkers)
	//fmt.Println("Number of Comparison Workers: ", *cWorkers)
	//fmt.Println("Input path: ", input)

	trees := make([]Tree, 0)

	readInput(&trees, input)
	//fmt.Println(trees, len(trees))

	// Compute hashes
	// hashMap is a map from hash to slice of tree indices
	// hashes is the array of hashes by index
	hashMap := make(map[uint32][]int, len(trees))
	hashes := make([]uint32, len(trees))

	// matrix is the adjacency matrix, initialized to false
	matrix := make([][]bool, len(trees))

	// Construct adjacency matrix
	for i := range matrix {
		matrix[i] = make([]bool, len(trees))
	}

	var lock sync.Mutex

  runtime.GC()
	// Start timing
	beginTime := time.Now()

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

  if hWorkers > len(trees) {
	  hWorkers = len(trees)
	}
	finishedHashMap := make(chan int, hWorkers)
  finishedHashTimer := make(chan int, hWorkers)

	treesPerWorker := len(trees) / hWorkers

  // Create dummy hash slice to time hashing algorithm
  dummyHashes := make([]uint32, len(trees))
  startHashing := time.Now()
  if !lockVar {
    hashChan := make(chan *[]uint32, hWorkers)

		for i := 0; i < hWorkers; i++ {
		  curTrees := trees[treesPerWorker*i : treesPerWorker*(i+1)]
			//fmt.Println(curTrees)
			go computeHashesParallel(&curTrees, treesPerWorker*i, hashChan)
		}

		for i := 0; i < hWorkers; i++ {
			hash := <-hashChan
			dummyHashes = append(dummyHashes, *hash...)
		}
  } else {
    for i := 0; i < hWorkers; i++ {
      curTrees := trees[treesPerWorker*i : treesPerWorker*(i+1)]
      go computeHashes(&curTrees, treesPerWorker*i, &dummyHashes, finishedHashTimer)
    }
    for i := 0; i < hWorkers; i++ {
      <-finishedHashTimer
    }
  }
  endHashing := time.Now()
  hashingTime := endHashing.Sub(startHashing).Nanoseconds()
  fmt.Println(hashingTime)

  // Actually compute the hashing and insert in the map
  beginHashingPlusInsert := time.Now()
	if !lockVar {
		// Start receiving hashes to insert in the map
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
				curTrees := trees[treesPerWorker*i : treesPerWorker*(i+1)]
				go computeHashesLock(&curTrees, treesPerWorker*i, &hashes, &hashMap, &lock, finishedHashMap)
			}
			for i := 0; i < hWorkers; i++ {
				<-finishedHashMap
			}
		} else {
			hashChan := make(chan *[]uint32, hWorkers)

			for i := 0; i < hWorkers; i++ {
				curTrees := trees[treesPerWorker*i : treesPerWorker*(i+1)]
				//fmt.Println(curTrees)
				go computeHashesSingle(&curTrees, treesPerWorker*i, hashChan, pairChan)
			}

			for i := 0; i < hWorkers; i++ {
				hash := <-hashChan
				hashes = append(hashes, *hash...)
			}
			// Pull from the 'finished' channel
			<-finishedHashMap
		}
	}
  endHashingPlusInsert := time.Now()
  hashingPlusInsertTime := endHashingPlusInsert.Sub(beginHashingPlusInsert).Nanoseconds()
  fmt.Println(hashingPlusInsertTime);

	//fmt.Println(hashMap)

	for _, list := range hashMap {
		//fmt.Println(list)
		for i := range list {
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
	for i := range matrix {
		matrix[i][i] = true
	}
	/*
	  for i := range matrix {
	    for j := 0; j < len(matrix[0]) - i; j++ {
	      result := Same(&trees[i], &trees[j], hashes[i], j, &hashMap)

	      // Mirror result to cut down on computation
	      matrix[i][j] = result
	      matrix[j][i] = result
	    }
	  }
	*/

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

// Walk walks the tree t sending all values
// from the tree to the channel ch.
func Walk(t *Tree, ch chan int) {
	_walk(t, ch)
	close(ch)
}

// Same determines whether the trees
// t1 and t2 contain the same values.
// Uses the hashes map to cheat
func Same(t1, t2 *Tree, hash1 uint32, i2 int, hashMap *map[uint32][]int) bool {
	equalTrees := (*hashMap)[hash1]
	for i := range equalTrees {
		if equalTrees[i] == i2 {
			return SameTraverse(t1, t2)
		}
	}
	//fmt.Println("t1: ", *t1, "\nt2: ", *t2, "\neT: ", equalTrees, "\n")
	return false
}

func SameTraverse(t1, t2 *Tree) bool {
	c1 := make(chan int)
	c2 := make(chan int)

	go Walk(t1, c1)
	go Walk(t2, c2)

	for {
		v1, ok1 := <-c1
		v2, ok2 := <-c2

		if v1 != v2 || ok1 != ok2 {
			return false
		}

		if !ok1 {
			break
		}
	}
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
