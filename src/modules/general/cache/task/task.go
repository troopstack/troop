package task

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/troopstack/troop/src/model"
)

type TaskScouts struct {
	TaskId        string
	Ch            chan int `json:"-"`
	Wg            sync.WaitGroup
	Detach        bool
	AcceptCount   int
	CompleteCount int
	Lock          bool
	M             map[string]*model.TaskScoutInfo
	CreateAt      time.Time
}

type SafeTasks struct {
	sync.RWMutex
	M map[string]*TaskScouts
}

var Tasks = NewSafeTasks()

func NewSafeTasks() *SafeTasks {
	return &SafeTasks{M: make(map[string]*TaskScouts)}
}

func (this *SafeTasks) PutTask(req *TaskScouts) {
	if _, exists := this.GetTask(req.TaskId); !exists {
		this.Lock()
		this.M[req.TaskId] = req
		this.Unlock()
	}
}

func (this *SafeTasks) PutTaskLock(taskId string, lock bool) {
	if task, exists := this.GetTask(taskId); exists {
		this.Lock()
		task.Lock = lock
		this.Unlock()
	}
}

func (this *SafeTasks) PutTaskTimeout(taskId string) {
	if task, exists := this.GetTask(taskId); exists {
		this.Lock()
		for scout := range task.M {
			if task.M[scout].Status == "wait" {
				task.M[scout].Status = "timeout"
				task.M[scout].Error = "timeout"
			}
		}
		this.Unlock()
	}
}

func (t *TaskScouts) TaskWait() {
	// 阻塞等待所有任务完成
	t.Wg.Wait()
	t.Ch <- 1
}

func (t *TaskScouts) TaskDone() {
	go func() {
		defer t.Wg.Done()
	}()
}

func (this *SafeTasks) GetTask(taskId string) (*TaskScouts, bool) {
	this.RLock()
	defer this.RUnlock()
	val, exists := this.M[taskId]
	return val, exists
}

func (this *SafeTasks) DeleteTask(taskId string) {
	this.Lock()
	defer this.Unlock()
	delete(this.M, taskId)
}

func (this *SafeTasks) TaskKeys() []string {
	this.RLock()
	defer this.RUnlock()
	count := len(this.M)
	keys := make([]string, count)
	i := 0
	for taskId := range this.M {
		keys[i] = taskId
		i++
	}
	return keys
}

func (this *SafeTasks) CreateTaskScout(req *model.TaskScoutInfo) {
	this.Lock()
	defer this.Unlock()
	if val, exists := this.M[req.TaskId]; exists {
		if _, scoutValExists := val.M[req.Scout]; !scoutValExists {
			if req.Status == "execution" {
				val.AcceptCount++
			} else if req.Status == "successful" {
				val.CompleteCount++
				val.TaskDone()
			} else if req.Status == "failed" {
				val.CompleteCount++
				val.TaskDone()
			} else if req.Status == "unreachable" {
				val.AcceptCount++
				val.CompleteCount++
				val.TaskDone()
			}
			val.M[req.Scout] = req
		}
	}
}

func (this *SafeTasks) PutTaskScoutStatus(taskId string, scout string, status string) {
	this.Lock()
	defer this.Unlock()
	if val, exists := this.M[taskId]; exists {
		if _, scoutValExists := val.M[scout]; scoutValExists {
			if val.M[scout].Status == "wait" && status == "execution" {
				val.AcceptCount++
			} else if val.M[scout].Status != "successful" && status == "successful" {
				val.CompleteCount++
				val.TaskDone()
			} else if val.M[scout].Status != "failed" && status == "failed" {
				val.CompleteCount++
				val.TaskDone()
			} else if val.M[scout].Status != "unreachable" && status == "unreachable" {
				val.AcceptCount++
				val.CompleteCount++
				val.TaskDone()
			}
			val.M[scout].Status = status
		}
	}
}

func (this *SafeTasks) PutTaskScoutResult(taskId string, scout string, res string, error bool) {
	this.Lock()
	defer this.Unlock()
	if val, exists := this.M[taskId]; exists {
		if _, scoutValExists := val.M[scout]; scoutValExists {
			if error {
				val.M[scout].Error += res
			} else {
				val.M[scout].Result += res
			}
		}
	}
}

func (this *SafeTasks) PutTaskScout(req *model.TaskScoutInfo) {
	this.Lock()
	defer this.Unlock()
	if val, exists := this.M[req.TaskId]; exists {
		if _, scoutValExists := val.M[req.Scout]; scoutValExists {
			val.M[req.Scout].Result += req.Result
			val.M[req.Scout].Error += req.Error
			if val.M[req.Scout].Status == "wait" && req.Status == "execution" {
				val.AcceptCount++
			} else if val.M[req.Scout].Status != "successful" && req.Status == "successful" {
				val.CompleteCount++
				val.TaskDone()
			} else if val.M[req.Scout].Status != "failed" && req.Status == "failed" {
				val.CompleteCount++
				val.TaskDone()
			} else if val.M[req.Scout].Status != "unreachable" && req.Status == "unreachable" {
				val.AcceptCount++
				val.CompleteCount++
				val.TaskDone()
			}
			val.M[req.Scout].Status = req.Status
		}
	}
}

func (this *SafeTasks) GetTaskScout(taskId string, scout string) (*model.TaskScoutInfo, bool) {
	this.RLock()
	defer this.RUnlock()
	val, exists := this.M[taskId]
	if exists {
		TaskScouts, exists := val.M[scout]
		return TaskScouts, exists
	}
	return nil, exists
}

var taskDataFile = "TASKS_DATA"

func SaveTasksToLocal() {
	b, err := json.Marshal(Tasks.M)
	if err != nil {
		log.Println("save tasks to local failed:", err.Error())
		return
	}
	err = ioutil.WriteFile(taskDataFile, b, 0666)
	if err != nil {
		log.Println("save tasks to local failed:", err.Error())
	}
}

func ReadTasksFromLocal() {
	Tasks.Lock()
	defer func() {
		Tasks.Unlock()
		os.Remove(taskDataFile)
	}()
	data, err := ioutil.ReadFile(taskDataFile)
	if data != nil {
		err = json.Unmarshal(data, &Tasks.M)
		if err != nil {
			log.Println("read tasks from local failed:", err.Error())
			return
		}
		log.Printf("Read %d tasks from local", len(Tasks.M))
	}

}
