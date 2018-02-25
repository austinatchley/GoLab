package main

import (
  "time"
  "strconv"
  "flag"
  "errors"
  "fmt"
  "strings"
  "bufio"
  "os"
)

// Structs
type Tree struct {
	Left  *Tree
  Value int
  Right *Tree
}

// Struct Methods
func (tree *Tree) AddNode(val int) (e error) {
  e = nil

  switch{
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
  go Walk(tree, c)

  var hash uint32 = 0

  // Unpack values from the channel c while it is OK
  val, ok := <-c
  for ; ok; val, ok = <-c {
    hash = hash + uint32(val) * 3433
  }
  return hash
}

// Main
func main() {
  hWorkers := flag.Int("hash-workers", 1, "number of workers on hashing")
  dWorkers := flag.Int("data-workers", 1, "number of workers on data")
  cWorkers := flag.Int("comp-workers", 1, "number of workers on comparison")

  var input string
  flag.StringVar(&input, "input", "", "path to input file")

  flag.Parse()

  _,_,_,_ = hWorkers, dWorkers, cWorkers, input

  //fmt.Println("Number of Hash Workers: ", *hWorkers)
  //fmt.Println("Number of Data Workers: ", *dWorkers)
  //fmt.Println("Number of Comparison Workers: ", *cWorkers)
  //fmt.Println("Input path: ", input)

  trees := make([]Tree, 1)

  readInput(&trees, input)
  // fmt.Println(trees, len(trees))

  beginTime := time.Now()

  // Compute hashes
  // hashMap is a map from hash to slice of Trees
  // hashes is the array of hashes by index
  hashMap := make(map[uint32][]int, len(trees))
  hashes := make([]uint32, len(trees))
  computeHashes(&trees, &hashes, &hashMap)

  // Construct adjacency matrix and compare trees
  matrix := make([][]bool, len(trees))
  for i := range matrix {
    matrix[i] = make([]bool, len(trees))
  }

  for _, list := range hashMap {
    for i := range list {
      for j := i; j < len(list); j++ {
        result := Same(&trees[i], &trees[j], hashes[i], j, &hashMap)

        // Mirror result to cut down on computation
        matrix[i][j] = result
        matrix[j][i] = result
      }
    }
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

  //printMatrix(&matrix)

  endTime := time.Now()
  diff := endTime.Sub(beginTime).Nanoseconds()
  fmt.Println(diff)
}

// Functions
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
  //fmt.Println("THESE SHOULD BE EQUAL\nt1: ", *t1, "\nt2: ", *t2, "\n")
  c1 := make(chan int)
  c2 := make(chan int)

  go Walk(t1, c1)
  go Walk(t2, c2)

  for {
    v1, ok1 := <-c1
    v2, ok2 := <-c2

    if v1 != v2 || ok1 != ok2{
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

func computeHashes(trees *[]Tree, hashes *[]uint32, hashMap *map[uint32][]int) {
  for i, elem := range *trees {
    hash := elem.Hash()
    (*hashes)[i] = hash
    (*hashMap)[hash] = append((*hashMap)[hash], i)
  }
}

// Utility Functions
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
  rawData, err := reader.ReadString('\n');
  for ; err == nil; rawData, err = reader.ReadString('\n') {
    check(err)

    strData := strings.Fields(rawData)
    data := make([]int, len(strData))
    for i, elem := range strData {
      data[i], err = strconv.Atoi(elem)
      check(err)
    }

    tree,err := createTree(&data)
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
