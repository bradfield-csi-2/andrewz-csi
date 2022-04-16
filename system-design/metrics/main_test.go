package main

import (
  "testing"
  . "sort"
  //"sort"
)

func BenchmarkSortInt64K(b *testing.B) {

	b.StopTimer()

	for i := 0; i < b.N; i++ {

		data := make([]int, 1<<16)

		for i := 0; i < len(data); i++ {

			data[i] = i ^ 0xcccc

		}

		b.StartTimer()

		Ints(data)

		b.StopTimer()

	}

}


func BenchmarkSortInt64K_Slice(b *testing.B) {

	b.StopTimer()

	for i := 0; i < b.N; i++ {

		data := make([]int, 1<<16)

		for i := 0; i < len(data); i++ {

			data[i] = i ^ 0xcccc

		}

		b.StartTimer()

		Slice(data, func(i, j int) bool { return data[i] < data[j] })

		b.StopTimer()

	}

}


func BenchmarkStableInt64K(b *testing.B) {

	b.StopTimer()

	for i := 0; i < b.N; i++ {

		data := make([]int, 1<<16)

		for i := 0; i < len(data); i++ {

			data[i] = i ^ 0xcccc

		}

		b.StartTimer()

		Stable(IntSlice(data))

		b.StopTimer()

	}

}
