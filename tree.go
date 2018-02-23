package main

import (
  "strconv"
  "flag"
  "errors"
  "fmt"
  "strings"
  "bufio"
  "os"
)

type Tree struct {
	Left  *Tree
  Value int
  Right *Tree
}

func (tree Tree) AddNode(val int) (e error) {
  e = nil

  if val < tree.Value {
    if tree.Left != nil {
      tree.Left.AddNode(val)
    } else {
      left := Tree{nil, val, nil}
      tree.Left = &left
    }
  } else if val > tree.Value {
    if tree.Right != nil {
      tree.Right.AddNode(val)
    } else {
      right := Tree{nil, val, nil}
      tree.Right = &right
    }
  } else if val == tree.Value {
    e = errors.New("given value is equal to the value of the current node")
  }
  return
}

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
func Same(t1, t2 *Tree) bool {
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

func check(e error) {
  if e != nil {
    panic(e)
  }
}

func createTree(data []int) (tree Tree, e error) {
  if len(data) == 0 {
    e = errors.New("length of data is 0")
    return
  }

  tree = Tree{nil, data[0], nil}
  for i := 1; i < len(data); i++ {
    err := tree.AddNode(data[i])
    check(err)
  }
  return
}

func main() { 
  hWorkers := flag.Int("hash-workers", 1, "number of workers on hashing")
  dWorkers := flag.Int("data-workers", 1, "number of workers on data")
  cWorkers := flag.Int("comp-workers", 1, "number of workers on comparison")
  
  //input := flag.String("input", "", "path to input file")
  var input string
  flag.StringVar(&input, "input", "", "path to input file")

  flag.Parse()
  
  fmt.Println("Number of Hash Workers: ", *hWorkers)
  fmt.Println("Number of Data Workers: ", *dWorkers)
  fmt.Println("Number of Comparison Workers: ", *cWorkers)
  fmt.Println("Input path: ", input)

  trees := make([]Tree, 10)
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

    tree,err := createTree(data)
    check(err)
    trees[index] = tree
    index++

    fmt.Println(data, len(data))
  }


}
