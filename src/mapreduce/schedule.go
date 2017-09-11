package mapreduce

import "sync"

//
// schedule() starts and waits for all tasks in the given phase (Map
// or Reduce). the mapFiles argument holds the names of the files that
// are the inputs to the map phase, one per map task. nReduce is the
// number of reduce tasks. the registerChan argument yields a stream
// of registered workers; each item is the worker's RPC address,
// suitable for passing to call(). registerChan will yield all
// existing registered workers (if any) and new ones as they register.
//
var wg sync.WaitGroup

func schedule(jobName string, mapFiles []string, nReduce int, phase jobPhase, registerChan chan string) {
	var ntasks int  //任务总数
	var n_other int // number of inputs (for reduce) or outputs (for map)
	switch phase {
	case mapPhase:
		ntasks = len(mapFiles)
		n_other = nReduce
	case reducePhase:
		ntasks = nReduce
		n_other = len(mapFiles)
	}

	MyLog.Printf("Schedule: %v %v tasks (%d I/Os)\n", ntasks, phase, n_other)

	//workers资源池
	workers := make(chan string, ntasks)
	go func() {
		for workerAdress := range registerChan {
			workers <- workerAdress
		}
		//原有错误代码
		// workerAdress := <-registerChan
		// workers <- workerAdress
	}()

	wg.Add(ntasks)
	for index := 0; index < ntasks; index++ {
		doTaskArgs := DoTaskArgs{
			JobName:       jobName,
			File:          mapFiles[index],
			Phase:         phase,
			TaskNumber:    index,
			NumOtherPhase: n_other,
		}
		go work(workers, doTaskArgs)
	}
	wg.Wait()
	// All ntasks tasks have to be scheduled on workers, and only once all of
	// them have been completed successfully should the function return.
	// Remember that workers may fail, and that any given worker may finish
	// multiple tasks.
	
	MyLog.Printf("Schedule: %v phase done\n", phase)
}

func work(workers chan string, doTaskArgs DoTaskArgs) {
	defer wg.Done()

	Here:
	workerAdress := <-workers //获取工作线程
	if call(workerAdress, "Worker.DoTask", doTaskArgs, nil) {//工作正常完成
		workers <- workerAdress 
	}else{//工作非正常完成
		// workers <- workerAdress 
		goto Here
	}
}
