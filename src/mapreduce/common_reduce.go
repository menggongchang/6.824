package mapreduce

import (
	"encoding/json"
	"log"
	"os"
	"sort"
)

type ByKey []KeyValue

func (keyValues ByKey) Len() int {
	return len(keyValues)
}
func (keyValues ByKey) Swap(i, j int) {
	keyValues[i], keyValues[j] = keyValues[j], keyValues[i]
}
func (keyValues ByKey) Less(i, j int) bool {
	return keyValues[i].Key < keyValues[j].Key
}

// doReduce manages one reduce task: it reads the intermediate
// key/value pairs (produced by the map phase) for this task, sorts the
// intermediate key/value pairs by key, calls the user-defined reduce function
// (reduceF) for each key, and writes the output to disk.
func doReduce(
	jobName string, // the name of the whole MapReduce job
	reduceTaskNumber int, // which reduce task this is
	outFile string, // write the output here
	nMap int, // the number of map tasks that were run ("M" in the paper)
	reduceF func(key string, values []string) string,
) {
	//
	// You will need to write this function.
	//
	// You'll need to read one intermediate file from each map task;
	// reduceName(jobName, m, reduceTaskNumber) yields the file
	// name from map task m.

	//1:读取每个map任务相应的输出文件，并将数据整合
	kvs := make([]KeyValue, 0)
	for index := 0; index < nMap; index++ {
		fileName := reduceName(jobName, index, reduceTaskNumber)
		file, err := os.Open(fileName)
		if err != nil {
			log.Fatal("doReduce: ", err)
		}
		dec := json.NewDecoder(file)
		for {
			var kv KeyValue
			err = dec.Decode(&kv)
			if err != nil {
				break
			}
			kvs = append(kvs, kv)
		}
		file.Close()
	}

	//2:将keyValues数据按照key值排序
	sort.Sort(ByKey(kvs))

	//3:每个key值对应相应的value切片
	keyValues := make(map[string][]string)
	for _, kv := range kvs {
		keyValues[kv.Key] = append(keyValues[kv.Key], kv.Value)
	}

	//4:写入相应的输出文件
	file, err := os.Create(outFile)
	if err != nil {
		log.Fatal(err)
	}
	enc := json.NewEncoder(file)
	kv := KeyValue{}
	for k, values := range keyValues {
		kv.Key = k
		kv.Value = reduceF(k, values)
		enc.Encode(&kv)
	}
	defer file.Close()

	// Your doMap() encoded the key/value pairs in the intermediate
	// files, so you will need to decode them. If you used JSON, you can
	// read and decode by creating a decoder and repeatedly calling
	// .Decode(&kv) on it until it returns an error.
	//
	// You may find the first example in the golang sort package
	// documentation useful.
	//
	// reduceF() is the application's reduce function. You should
	// call it once per distinct key, with a slice of all the values
	// for that key. reduceF() returns the reduced value for that key.
	//
	// You should write the reduce output as JSON encoded KeyValue
	// objects to the file named outFile. We require you to use JSON
	// because that is what the merger than combines the output
	// from all the reduce tasks expects. There is nothing special about
	// JSON -- it is just the marshalling format we chose to use. Your
	// output code will look something like this:
	//
	// enc := json.NewEncoder(file)
	// for key := ... {
	// 	enc.Encode(KeyValue{key, reduceF(...)})
	// }
	// file.Close()
	//
}
