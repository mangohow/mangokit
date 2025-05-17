package stream

import (
	"reflect"
	"strconv"
	"testing"
)

func TestMap(t *testing.T) {
	t.Run("primitive type conversion", func(t *testing.T) {
		in := []int{1, 2, 3}
		want := []string{"1", "2", "3"}
		got := Map(in, func(n int) string {
			return string(rune(n + 48)) // ASCII 转换
		})
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Map() = %v, want %v", got, want)
		}
	})

	t.Run("struct to value", func(t *testing.T) {
		type person struct{ Age int }
		in := []person{{20}, {30}, {40}}
		want := []int{20, 30, 40}
		got := Map(in, func(p person) int {
			return p.Age
		})
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Map() = %v, want %v", got, want)
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		var in []int
		got := Map(in, func(n int) string { return "" })
		if len(got) != 0 {
			t.Errorf("Map() should return empty slice")
		}
	})

	t.Run("nil slice", func(t *testing.T) {
		var in []float64
		got := Map(in, func(f float64) bool { return true })
		if got != nil {
			t.Errorf("Map() should return nil for nil input")
		}
	})

}

func TestMapP(t *testing.T) {
	t.Run("basic struct field access", func(t *testing.T) {
		// 验证通过指针访问结构体字段
		type Data struct{ X int }
		in := []Data{{1}, {2}, {3}}
		want := []int{1, 2, 3}
		got := MapP(in, func(d *Data) int {
			return d.X // 通过指针访问字段
		})
		if !reflect.DeepEqual(got, want) {
			t.Errorf("MapP() = %v, want %v", got, want)
		}
	})

	t.Run("modify original data via pointer", func(t *testing.T) {
		// 验证指针参数可以修改原始数据
		in := []int{1, 2, 3}
		copySlice := make([]int, len(in))
		copy(copySlice, in)

		// 执行映射但不修改（确保测试独立性）
		MapP(in, func(n *int) int {
			return *n * 2
		})

		// 验证原始数据未被意外修改
		if !reflect.DeepEqual(in, copySlice) {
			t.Errorf("MapP should not modify original data, got %v want %v", in, copySlice)
		}
	})

	t.Run("nil slice handling", func(t *testing.T) {
		var in []complex128
		got := MapP(in, func(c *complex128) bool { return true })
		if got != nil {
			t.Error("MapP should return nil for nil input")
		}
	})

	t.Run("zero value elements", func(t *testing.T) {
		in := make([]string, 3) // ["", "", ""]
		want := []int{0, 0, 0}
		got := MapP(in, func(s *string) int {
			return len(*s)
		})
		if !reflect.DeepEqual(got, want) {
			t.Errorf("MapP() = %v, want %v", got, want)
		}
	})

}

func TestFilter(t *testing.T) {
	t.Run("basic filtering", func(t *testing.T) {
		in := []int{1, 2, 3, 4, 5}
		want := []int{2, 4}
		got := Filter(in, func(n int) bool {
			return n%2 == 0
		})
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Filter() = %v, want %v", got, want)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		var in []string
		got := Filter(in, func(s string) bool { return true })
		if got != nil {
			t.Error("Filter() should return nil for empty slice")
		}
	})

	t.Run("all elements filtered out", func(t *testing.T) {
		in := []float64{1.1, 2.2, 3.3}
		got := Filter(in, func(f float64) bool { return f > 5.0 })
		if len(got) != 0 {
			t.Error("Filter() should return empty slice when no elements match")
		}
	})

	t.Run("struct elements", func(t *testing.T) {
		type Data struct{ Valid bool }
		in := []Data{{true}, {false}, {true}}
		want := []Data{{true}, {true}}
		got := Filter(in, func(d Data) bool { return d.Valid })
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Filter() = %v, want %v", got, want)
		}
	})

}

func TestFilterP(t *testing.T) {
	t.Run("pointer access optimization", func(t *testing.T) {
		type BigStruct struct{ data [1024]byte }
		in := make([]BigStruct, 10)
		got := FilterP(in, func(bs *BigStruct) bool {
			return bs.data[0] == 0 // 通过指针访问大结构体
		})
		if len(got) != 10 {
			t.Error("FilterP failed to process large structs")
		}
	})

	t.Run("modify original data", func(t *testing.T) {
		in := []int{1, 2, 3}
		original := make([]int, len(in))
		copy(original, in)

		// 测试是否意外修改原数据
		FilterP(in, func(n *int) bool {
			*n *= 2 // 故意修改原数据
			return *n > 2
		})

		// 验证原数据是否被修改
		if !reflect.DeepEqual(in, []int{2, 4, 6}) {
			t.Errorf("FilterP may modify original data: got %v", in)
		}
	})

	t.Run("zero value elements", func(t *testing.T) {
		in := make([]string, 3) // ["", "", ""]
		got := FilterP(in, func(s *string) bool {
			return *s == ""
		})
		if len(got) != 3 {
			t.Errorf("FilterP() = %v, want 3 elements", got)
		}
	})

	t.Run("nil slice handling", func(t *testing.T) {
		var in []complex64
		got := FilterP(in, func(c *complex64) bool { return true })
		if got != nil {
			t.Error("FilterP should return nil for nil input")
		}
	})

	t.Run("pointer type elements", func(t *testing.T) {
		in := []*int{new(int), nil}
		got := FilterP(in, func(p **int) bool {
			return *p != nil
		})
		if len(got) != 1 {
			t.Error("FilterP should filter pointer elements correctly")
		}
	})
}

func TestForEach(t *testing.T) {
	t.Run("process all elements", func(t *testing.T) {
		counter := 0
		ForEach([]int{1, 2, 3}, func(n int) bool {
			counter++
			return true
		})
		if counter != 3 {
			t.Errorf("Processed %d elements, want 3", counter)
		}
	})

	t.Run("stop on first false", func(t *testing.T) {
		processed := []string{}
		ForEach([]string{"a", "b", "c"}, func(s string) bool {
			processed = append(processed, s)
			return s != "b"
		})
		if !reflect.DeepEqual(processed, []string{"a", "b"}) {
			t.Errorf("Processed %v, want [a b]", processed)
		}
	})

	t.Run("empty slice no-op", func(t *testing.T) {
		called := false
		ForEach([]int{}, func(int) bool {
			called = true
			return true
		})
		if called {
			t.Error("Unexpected call on empty slice")
		}
	})

	t.Run("nil slice safety", func(t *testing.T) {
		var nilSlice []float64
		ForEach(nilSlice, func(f float64) bool {
			t.Fatal("Should not process nil slice")
			return true
		})
	})
}

func TestForEachP(t *testing.T) {
	t.Run("modify through pointer", func(t *testing.T) {
		data := []int{1, 2, 3}
		ForEachP(data, func(n *int) bool {
			*n *= 2
			return true
		})
		if !reflect.DeepEqual(data, []int{2, 4, 6}) {
			t.Errorf("Modified data = %v, want [2 4 6]", data)
		}
	})

	t.Run("partial processing", func(t *testing.T) {
		modifications := 0
		data := []int{1, 2, 3, 4}
		ForEachP(data, func(n *int) bool {
			if *n == 3 {
				return false
			}
			*n *= 10
			modifications++
			return true
		})
		if modifications != 2 || data[0] != 10 || data[1] != 20 || data[2] != 3 {
			t.Errorf("Partial modify result = %v (mods: %d)", data, modifications)
		}
	})

	t.Run("handle nil elements", func(t *testing.T) {
		var slice []*int
		slice = append(slice, nil, new(int))
		nilCount := 0
		ForEachP(slice, func(p **int) bool {
			if *p == nil {
				nilCount++
			}
			return true
		})
		if nilCount != 1 {
			t.Errorf("Detected %d nil pointers, want 1", nilCount)
		}
	})

	t.Run("struct field access", func(t *testing.T) {
		type Record struct{ ID int }
		records := []Record{{1}, {2}, {3}}
		sum := 0
		ForEachP(records, func(r *Record) bool {
			sum += r.ID
			return true
		})
		if sum != 6 {
			t.Errorf("Sum of IDs = %d, want 6", sum)
		}
	})

	t.Run("zero-capacity slice", func(t *testing.T) {
		s := make([]int, 0)
		ForEachP(s, func(n *int) bool {
			t.Error("Should not process zero-capacity slice")
			return true
		})
	})
}

func TestDelete(t *testing.T) {
	t.Run("delete middle element", func(t *testing.T) {
		s := []int{1, 2, 3, 4}
		got := Delete(s, 1)
		want := []int{1, 3, 4}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Delete() = %v, want %v", got, want)
		}
	})

	t.Run("delete first element", func(t *testing.T) {
		s := []string{"a", "b", "c"}
		got := Delete(s, 0)
		want := []string{"b", "c"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Delete() = %v, want %v", got, want)
		}
	})

	t.Run("delete last element", func(t *testing.T) {
		s := []float64{1.1, 2.2, 3.3}
		got := Delete(s, 2)
		want := []float64{1.1, 2.2}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Delete() = %v, want %v", got, want)
		}
	})

	t.Run("panic on out of range", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Delete() did not panic on invalid index")
			}
		}()
		Delete([]int{1}, 1) // 索引越界
	})

	t.Run("empty slice handling", func(t *testing.T) {
		var s []int
		got := Delete(s, 0)
		if len(got) != 0 {
			t.Error("Delete() should return empty slice")
		}
	})
}

func TestDeleteFunc(t *testing.T) {
	t.Run("delete all matches", func(t *testing.T) {
		s := []int{2, 4, 6, 8}
		got := DeleteFunc(s, func(n int) bool { return n%2 == 0 })
		if len(got) != 0 {
			t.Errorf("DeleteFunc() = %v, want empty slice", got)
		}
	})

	t.Run("delete partial elements", func(t *testing.T) {
		type Data struct{ Valid bool }
		s := []Data{{true}, {false}, {true}}
		got := DeleteFunc(s, func(d Data) bool { return d.Valid })
		if len(got) != 1 || got[0].Valid {
			t.Errorf("DeleteFunc() = %v, want [false]", got)
		}
	})

	t.Run("no elements deleted", func(t *testing.T) {
		s := []string{"go", "rust", "zig"}
		got := DeleteFunc(s, func(s string) bool { return false })
		if len(got) != 3 {
			t.Errorf("DeleteFunc() = %v, want original slice", got)
		}
	})

	t.Run("clear trailing elements", func(t *testing.T) {
		s := make([]*int, 3)
		s[0] = new(int)
		got := DeleteFunc(s, func(p *int) bool { return p == nil })
		if len(got) != 1 || got[0] != s[0] {
			t.Error("DeleteFunc() failed to clear unused pointers")
		}
	})

	t.Run("capacity preservation", func(t *testing.T) {
		s := make([]int, 4, 10)
		got := DeleteFunc(s, func(n int) bool { return true })
		if cap(got) != 10 {
			t.Errorf("DeleteFunc() changed capacity to %d", cap(got))
		}
	})
}

func TestEvery(t *testing.T) {
	t.Run("all elements satisfy condition", func(t *testing.T) {
		nums := []int{2, 4, 6, 8}
		got := Every(nums, func(n int) bool { return n%2 == 0 })
		if !got {
			t.Error("Every() = false, want true for all even numbers")
		}
	})

	t.Run("one element violates", func(t *testing.T) {
		words := []string{"apple", "banana", "cherry", "12"}
		got := Every(words, func(s string) bool { return len(s) > 3 })
		if got {
			t.Error("Every() = true, want false for short element")
		}
	})

	t.Run("empty slice returns true", func(t *testing.T) {
		var empty []float64
		got := Every(empty, func(f float64) bool { return f > 0 })
		if !got {
			t.Error("Every() = false, want true for empty slice")
		}
	})

	t.Run("stop on first failure", func(t *testing.T) {
		count := 0
		data := []int{1, 2, 3, 4}
		Every(data, func(n int) bool {
			count++
			return n < 3
		})
		if count != 3 {
			t.Errorf("Processed %d elements, want 2", count)
		}
	})

	t.Run("struct elements", func(t *testing.T) {
		type Item struct{ Valid bool }
		items := []Item{{true}, {true}, {true}}
		got := Every(items, func(i Item) bool { return i.Valid })
		if !got {
			t.Error("Every() failed for valid struct elements")
		}
	})
}

func TestSome(t *testing.T) {
	t.Run("has matching element", func(t *testing.T) {
		nums := []int{1, 3, 5, 7, 8}
		got := Some(nums, func(n int) bool { return n%2 == 0 })
		if !got {
			t.Error("Some() = false, want true for existing even")
		}
	})

	t.Run("no elements match", func(t *testing.T) {
		words := []string{"go", "rust", "zig"}
		got := Some(words, func(s string) bool { return len(s) > 5 })
		if got {
			t.Error("Some() = true, want false for all short strings")
		}
	})

	t.Run("empty slice returns false", func(t *testing.T) {
		var empty []*int
		got := Some(empty, func(p *int) bool { return p != nil })
		if got {
			t.Error("Some() = true, want false for empty slice")
		}
	})

	t.Run("stop on first success", func(t *testing.T) {
		count := 0
		data := []int{1, 2, 3, 4}
		Some(data, func(n int) bool {
			count++
			return n > 0
		})
		if count != 1 {
			t.Errorf("Processed %d elements, want 1", count)
		}
	})

	t.Run("pointer elements", func(t *testing.T) {
		var nilPtr *int
		ptrs := []*int{nilPtr, new(int), nil}
		got := Some(ptrs, func(p *int) bool { return p != nil })
		if !got {
			t.Error("Some() failed to find non-nil pointer")
		}
	})
}

func TestReduce(t *testing.T) {
	t.Run("sum integers", func(t *testing.T) {
		nums := []int{1, 2, 3, 4}
		got := Reduce(nums, 0, func(acc, n int) int { return acc + n })
		if got != 10 {
			t.Errorf("Sum Reduce() = %d, want 10", got)
		}
	})

	t.Run("string concatenation", func(t *testing.T) {
		words := []string{"go", " ", "1.21"}
		got := Reduce(words, "", func(acc, s string) string { return acc + s })
		if got != "go 1.21" {
			t.Errorf("Concat Reduce() = %q, want 'go 1.21'", got)
		}
	})

	t.Run("empty slice returns initial", func(t *testing.T) {
		var empty []float64
		got := Reduce(empty, 3.14, func(acc float64, _ float64) float64 { return 0 })
		if got != 3.14 {
			t.Error("Empty Reduce() modified initial value")
		}
	})

	t.Run("type conversion reduce", func(t *testing.T) {
		nums := []int{1, 2, 3}
		got := Reduce(nums, "", func(acc string, n int) string {
			return acc + strconv.Itoa(n)
		})
		if got != "123" {
			t.Errorf("Type conversion Reduce() = %q, want '123'", got)
		}
	})

	t.Run("struct accumulation", func(t *testing.T) {
		type Point struct{ X, Y int }
		points := []Point{{1, 2}, {3, 4}, {5, 6}}
		got := Reduce(points, Point{}, func(acc Point, p Point) Point {
			return Point{acc.X + p.X, acc.Y + p.Y}
		})
		if got.X != 9 || got.Y != 12 {
			t.Errorf("Struct Reduce() = %v, want {9 12}", got)
		}
	})

	t.Run("conditional accumulation", func(t *testing.T) {
		nums := []int{1, 2, 3, 4, 5}
		got := Reduce(nums, 0, func(acc, n int) int {
			if n%2 == 0 {
				return acc + n
			}
			return acc
		})
		if got != 6 {
			t.Errorf("Conditional Reduce() = %d, want 6", got)
		}
	})

	t.Run("nil slice handling", func(t *testing.T) {
		var nilSlice []int
		got := Reduce(nilSlice, 100, func(acc, _ int) int { return 0 })
		if got != 100 {
			t.Error("Nil Reduce() modified initial value")
		}
	})

	t.Run("order preservation", func(t *testing.T) {
		ops := []string{"A", "B", "C"}
		var log []string
		Reduce(ops, "", func(acc string, s string) string {
			log = append(log, s)
			return acc + s
		})
		if !reflect.DeepEqual(log, []string{"A", "B", "C"}) {
			t.Errorf("Processing order %v, want [A B C]", log)
		}
	})
}
