// +build property

package scheduler

import (
	"errors"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/scootdev/scoot/cloud/cluster"
	"github.com/scootdev/scoot/common/stats"
	"github.com/scootdev/scoot/saga"
	"github.com/scootdev/scoot/saga/sagalogs"
	"github.com/scootdev/scoot/sched"
	"github.com/scootdev/scoot/sched/worker"
	"github.com/scootdev/scoot/sched/worker/workers"
	"testing"
)

// objects needed to initialize a stateful scheduler
type schedulerDeps struct {
	initialCl []cluster.Node
	clUpdates chan []cluster.NodeUpdate
	sc        saga.SagaCoordinator
	wf        worker.WorkerFactory
}

// returns default scheduler deps populated with in memory fakes
// The default cluster has 5 nodes
func getDefaultSchedDeps() *schedulerDeps {
	cl := makeTestCluster("node1", "node2", "node3", "node4", "node5")
	return &schedulerDeps{
		initialCl: cl.nodes,
		clUpdates: cl.ch,
		sc:        sagalogs.MakeInMemorySagaCoordinator(),
		wf: func(cluster.Node) worker.Worker {
			return workers.MakeSimWorker()
		},
	}
}

func makeStatefulSchedulerDeps(deps *schedulerDeps) *statefulScheduler {

	return NewStatefulScheduler(
		deps.initialCl,
		deps.clUpdates,
		deps.sc,
		deps.wf,
		stats.NilStatsReceiver(),
		true)
}

func makeDefaultStatefulScheduler() *statefulScheduler {
	return makeStatefulSchedulerDeps(getDefaultSchedDeps())
}

// ensure a scheduler initializes to the correct state
func Test_StatefulScheduler_Initialize(t *testing.T) {

	s := makeDefaultStatefulScheduler()

	if len(s.inProgressJobs) != 0 {
		t.Errorf("Expected Scheduler to startup with no jobs in progress")
	}

	if len(s.clusterState.nodes) != 5 {
		t.Errorf("Expected Scheduler to have a cluster with 5 nodes")
	}
}

func Test_StatefulScheduler_ScheduleJobSuccess(t *testing.T) {
	jobDef := sched.GenJobDef(1)

	//mock sagalog
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	sagaLogMock := saga.NewMockSagaLog(mockCtrl)
	sagaLogMock.EXPECT().StartSaga(gomock.Any(), nil)

	deps := getDefaultSchedDeps()
	deps.sc = saga.MakeSagaCoordinator(sagaLogMock)
	s := makeStatefulSchedulerDeps(deps)

	id, err := s.ScheduleJob(jobDef)
	if id == "" {
		t.Errorf("Expected successfully scheduled job to return non empty job string!")
	}

	if err != nil {
		t.Errorf("Expected job to be Scheduled Successfully %v", err)
	}
}

func Test_StatefulScheduler_ScheduleJobFailure(t *testing.T) {
	jobDef := sched.GenJobDef(1)

	//mock sagalog
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	sagaLogMock := saga.NewMockSagaLog(mockCtrl)
	sagaLogMock.EXPECT().StartSaga(gomock.Any(), gomock.Any()).Return(errors.New("test error"))

	deps := getDefaultSchedDeps()
	deps.sc = saga.MakeSagaCoordinator(sagaLogMock)
	s := makeStatefulSchedulerDeps(deps)

	id, err := s.ScheduleJob(jobDef)
	if id != "" {
		t.Errorf("Expected unsuccessfully scheduled job to return an empty job string!")
	}

	if err == nil {
		t.Error("Expected job return error")
	}
}

func Test_StatefulScheduler_AddJob(t *testing.T) {
	s := makeDefaultStatefulScheduler()
	jobDef := sched.GenJobDef(1)
	id, _ := s.ScheduleJob(jobDef)

	// advance scheduler loop & then verify state
	s.step()
	if len(s.inProgressJobs) != 1 {
		t.Errorf("Expected In Progress Jobs to be 1 not %v", len(s.inProgressJobs))
	}

	_, ok := s.inProgressJobs[id]
	if !ok {
		t.Errorf("Expected the %v to be an inProgressJobs", id)
	}
}

// verify that jobs are distributed evenly
func Test_StatefulScheduler_TasksDistributedEvenly(t *testing.T) {
	jobDef := sched.GenJobDef(1000)
	s := makeDefaultStatefulScheduler()

	//initialize NodeMap to keep track of tasks per node
	taskMap := make(map[string]cluster.NodeId)

	/*jobId, _ :=*/ s.ScheduleJob(jobDef)
	s.step()

	for len(s.inProgressJobs) > 0 {
		s.step()

		for nodeId, state := range s.clusterState.nodes {
			if state.runningTask != noTask {
				taskMap[state.runningTask] = nodeId
			}
		}
	}

	taskCountMap := make(map[cluster.NodeId]int)
	for _, nodeId := range taskMap {
		taskCountMap[nodeId]++
	}

	// The in memory workers aren't doing anything interesting except sleeping distribution
	// should be even with in 180 - 220 nodes otherwise something is wrong.
	for nodeId, taskCount := range taskCountMap {
		if taskCount < 180 || taskCount > 220 {
			t.Fatalf(`Tasks were not evenly distributed across nodes.  Expected each node
				to have 190 to 210 tasks executed on it. %v had an unequal number of tasks %v scheduled 
				on it.  TaskCountMap: %+v`, nodeId, taskCount, taskCountMap)
		}
	}

	fmt.Printf("Task to Node Distribution: %+v", taskCountMap)
}

// Ensure a single job with one task runs to completion, updates
// state correctly, and makes the expected calls to the SagaLog
func Test_StatefulScheduler_JobRunsToCompletion(t *testing.T) {
	jobDef := sched.GenJobDef(1)
	var taskIds []string
	for taskId, _ := range jobDef.Tasks {
		taskIds = append(taskIds, taskId)
	}
	taskId := taskIds[0]

	deps := getDefaultSchedDeps()
	// cluster with one node
	cl := makeTestCluster("node1")
	deps.initialCl = cl.nodes
	deps.clUpdates = cl.ch

	// sagalog mock to ensure all messages are logged appropriately
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	sagaLogMock := saga.NewMockSagaLog(mockCtrl)
	sagaLogMock.EXPECT().StartSaga(gomock.Any(), gomock.Any())

	deps.sc = saga.MakeSagaCoordinator(sagaLogMock)

	s := makeStatefulSchedulerDeps(deps)

	// add job and run through scheduler
	jobId, _ := s.ScheduleJob(jobDef)

	// add additional saga data
	sagaLogMock.EXPECT().LogMessage(saga.MakeStartTaskMessage(jobId, taskId, nil))
	endMessageMatcher := TaskMessageMatcher{JobId: jobId, TaskId: taskId, Data: gomock.Any()}
	sagaLogMock.EXPECT().LogMessage(endMessageMatcher)
	sagaLogMock.EXPECT().LogMessage(saga.MakeEndSagaMessage(jobId))
	s.step()

	// advance scheduler verify task got added & scheduled
	for s.inProgressJobs[jobId].Tasks[taskId].Status == sched.NotStarted {
		s.step()
	}

	// verify scheduler state updated appropriately
	if s.clusterState.nodes["node1"].runningTask != taskId {
		t.Errorf("Expected %v to be scheduled on node1.  nodestate: %+v", taskId, s.clusterState.nodes["node1"])
	}

	// advance scheduler until the task completes
	for s.inProgressJobs[jobId].Tasks[taskId].Status == sched.InProgress {
		s.step()
	}

	// verify state changed appropriately
	if s.clusterState.nodes["node1"].runningTask != noTask {
		t.Errorf("Expected node1 to not have any running tasks")
	}

	// advance scheduler until job gets marked completed
	for s.inProgressJobs[jobId].getJobStatus() != sched.Completed {
		s.step()
	}

	// verify that EndSaga Message gets logged
	if !s.inProgressJobs[jobId].EndingSaga {
		t.Errorf("Expected Completed job to be EndingSaga")
	}

	for len(s.inProgressJobs) > 0 {
		s.step()
	}
}
