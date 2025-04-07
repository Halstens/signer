package main

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"sync"
)

func main() {
	Workers()
}

func Workers() {
	hashSignJobs := []job{
		job(DataProcessing),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
	}
	ExecutePipeline(hashSignJobs...)
}

// // Конвеерная обработка функций воркеров
func ExecutePipeline(works ...job) {
	in := make(chan interface{}, 100)
	wg := &sync.WaitGroup{}

	for _, work := range works {
		out := make(chan interface{}, 100)
		wg.Add(1)
		go func(work job, in, out chan interface{}) {
			defer close(out)
			defer wg.Done()
			work(in, out)

		}(work, in, out)
		in = out
	}
	wg.Wait()
}

func DataProcessing(in, out chan interface{}) {
	inputData := []int{0, 1}
	for _, fibNum := range inputData {
		out <- fibNum
		fmt.Print(fibNum)
	}
	fmt.Println("Data")
}

func SingleHash(in, out chan interface{}) {
	for val := range in {
		if inter, ok := val.(int); ok {
			var buf bytes.Buffer
			result := ""
			str := fmt.Sprintf("%d", inter)
			md5Chan := make(chan string, 1)
			crc32Chan := make(chan string, 1)
			crc32md5Chan := make(chan string, 1)

			go func(val chan string) {
				val <- DataSignerCrc32(str)
			}(crc32Chan)
			go func(valmd5, crcmd chan string) {
				md5 := DataSignerMd5(str)
				md5Chan <- md5
				crcmd <- DataSignerCrc32(md5)
			}(md5Chan, crc32md5Chan)

			crc32Data := <-crc32Chan
			md5Data := <-md5Chan
			crc32md5Data := <-crc32md5Chan

			// fmt.Println(str + " SingleHash data " + str)
			// fmt.Println(str + " SingleHash md5(data) " + md5Data)
			// fmt.Println(str + " SingleHash crc32(md5(data)) " +
			// 	crc32md5Data)
			// fmt.Println(str + " SingleHash crc32(data) " +
			// 	crc32Data)

			buf.WriteString(str + " SingleHash data " + str)
			buf.WriteString(str + " SingleHash md5(data) " + md5Data)
			buf.WriteString(str + " SingleHash crc32(md5(data)) " +
				crc32md5Data)
			buf.WriteString(str + " SingleHash crc32(data) " +
				crc32Data)
			result += crc32Data + "~" +
				crc32md5Data
			//fmt.Println(str + " SingleHash result " + result)
			buf.WriteString(str + " SingleHash result " + result)
			os.Stdout.Write(buf.Bytes())
			out <- result
		} else {
			fmt.Println("Нихуя не вышло")
		}
		fmt.Println()
	}

}

func MultiHash(in, out chan interface{}) {
	globWG := &sync.WaitGroup{}
	//var buf bytes.Buffer
	for val := range in {
		if str, ok := val.(string); ok {
			// горутина для каждого значения в in канале
			globWG.Add(1)
			go func(s string) {
				defer globWG.Done()
				result := ""
				wg := &sync.WaitGroup{}
				mu := &sync.Mutex{}
				resParts := make([]string, 6)
				// 6 горутин для каждого значения из входного канала
				for i := 0; i < 6; i++ {
					wg.Add(1)
					go func(indexNum int) {
						defer wg.Done()
						num := fmt.Sprintf("%d", indexNum)
						hashData := DataSignerCrc32(num + str)
						mu.Lock()
						resParts[indexNum] = hashData
						mu.Unlock()
						// buf.WriteString(str + " MultiHash: crc32(th+step1) " +
						// 	num + " " + hashData)
						// os.Stdout.Write(buf.Bytes())
						fmt.Println(str + " MultiHash: crc32(th+step1) " +
							num + " " + hashData)
					}(i)
				}
				wg.Wait()

				for _, part := range resParts {
					result += part
				}

				//buf.WriteString(str + " MultiHash result: " + result)
				fmt.Println(str + " MultiHash result: " + result)
				out <- result
			}(str)

		} else {
			fmt.Println("No OK")
		}

		globWG.Wait()

	}
	//os.Stdout.Write(buf.Bytes())

}

func CombineResults(in, out chan interface{}) {

	fmt.Println("CombineResult")
	var Results = make([]string, 2)
	FinalRes := ""
	for val := range in {
		if str, ok := val.(string); ok {
			Results = append(Results, str)
		} else {
			fmt.Println("Комбайн не прошел")
		}
	}

	sort.Slice(Results, func(i, j int) bool {
		return Results[i] < Results[j]
	})

	for id, item := range Results {
		if item != "" && id != len(Results)-1 {
			FinalRes += item + "_"
		} else {
			FinalRes += item
		}

	}

	out <- FinalRes

	fmt.Println("CombineResults " + FinalRes)

}
