package workloadplan

import (
	"context"
	v1 "github.com/gzlj/appplan/pkg/apis/plan/v1"
	"strconv"

	//v1 "example.com/gzlj/workloadplan/api/v1"
	//"example.com/gzlj/workloadplan/pkg/global"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	//"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"sync"
	"time"
)

var (
	Manager = NewPlanMgr()
)

type IPlan interface {
	Run()
	Pause()
	Destroy()
	Name() string
}


type Mgr struct {
	plans map[string]IPlan
	lock sync.RWMutex
}

func NewPlanMgr() *Mgr{
	return &Mgr{
		plans: make(map[string]IPlan),
	}
}

func (mgr *Mgr) AddOrUpdatePlan(p IPlan) (err error){
	if p == nil || p.Name() == ""{
		err = fmt.Errorf("Input Plan is not valied.")
		return
	}
	mgr.lock.Lock()
	defer mgr.lock.Unlock()
	old, ok := mgr.plans[p.Name()]
	if ok {
		old.Destroy()
	}
	mgr.plans[p.Name()] = p
	p.Run()
	log.Info("Add one plan.")
	return
}

func (mgr *Mgr) DeletePlan(name string) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()
	p, ok  := mgr.plans[name]
	if !ok {
		return
	}
	p.Destroy()
	delete(mgr.plans, name)
}

func (mgr *Mgr) PausePlan(name string) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()
	p, ok  := mgr.plans[name]
	if !ok {
		return
	}
	p.Pause()
}


type NormalPlan struct {
	*v1.WorkLoadPlan
	looping bool
	pause bool
	waitSeconds int
	ctx context.Context
	cancelFunc context.CancelFunc
}

func (p *NormalPlan) Name() string {
	if p.WorkLoadPlan == nil {
		return ""
	}
	return p.WorkLoadPlan.Namespace + "/" + p.WorkLoadPlan.Name
}

func (p *NormalPlan) Run() {
	if p.looping == false {
		p.looping = true
		go p.run(p.ctx)
	}
}


func (p *NormalPlan) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			//log.Info("run loop.")
			if p.pause {
				//log.Info("p.pause == true")
				continue
			}
			//sync workload
			//log.Info("p.syncWorkloadsReplicas()")
			p.syncWorkloadsReplicas()
		}
		time.Sleep(time.Second * time.Duration(p.waitSeconds))
	}
}

func (p *NormalPlan) syncWorkloadsReplicasByCron(cron *v1.CrondPlan, workloads []*v1.WorkLoad) {

	period := strings.ToUpper(cron.Period)
	//format := "2006-01-02 15:04:05"
	var (
		err error
		startTime time.Time
		endTime time.Time
		startHour, endHour  int
		startMinute, endMinute int
		startSecond, endSecond int
	)
	startHour, err = strconv.Atoi(cron.StartHour)
	if err != nil {
		return
	}
	endHour, err = strconv.Atoi(cron.EndHour)
	if err != nil {
		return
	}

	startMinute, err = strconv.Atoi(cron.StartMinute)
	if err != nil {
		return
	}
	endMinute, err = strconv.Atoi(cron.EndMinute)
	if err != nil {
		return
	}

	startSecond, err = strconv.Atoi(cron.StartSecond)
	if err != nil {
		return
	}

	endSecond, err = strconv.Atoi(cron.EndSecond)
	if err != nil {
		return
	}
	now := time.Now()
	switch period {
	case CRON_PERIOD_DAY:
		startTime = time.Date(now.Year(), now.Month(), now.Day(), startHour, startMinute, startSecond, 0, time.Local)
		endTime = time.Date(now.Year(), now.Month(), now.Day(), endHour, endMinute, endSecond, 0, time.Local)

	case CRON_PERIOD_MONTH:
		startDay, err := strconv.Atoi(cron.StartDay)
		if err != nil {
			return
		}
		endDay, err := strconv.Atoi(cron.EndDay)
		if err != nil {
			return
		}
		startTime = time.Date(now.Year(), now.Month(), startDay, startHour, startMinute, startSecond, 0, time.Local)
		endTime = time.Date(now.Year(), now.Month(), endDay, endHour, endMinute, endSecond, 0, time.Local)

	case CRON_PERIOD_YEAR:
		startDay, err := strconv.Atoi(cron.StartDay)
		if err != nil {
			return
		}
		endDay, err := strconv.Atoi(cron.EndDay)
		if err != nil {
			return
		}

		startMonth, err := strconv.Atoi(cron.StartMonth)
		if err != nil {
			return
		}
		endMonth, err := strconv.Atoi(cron.EndMonth)
		if err != nil {
			return
		}
		startTime = time.Date(now.Year(), time.Month(startMonth), startDay, startHour, startMinute, startSecond, 0, time.Local)
		endTime = time.Date(now.Year(), time.Month(endMonth), endDay, endHour, endMinute, endSecond, 0, time.Local)

	}

	if now.After(startTime) && now.Before(endTime) {
		p.syncReplicas(workloads)
	} else {
		p.setReplicasZero(workloads)
	}

}

func (p *NormalPlan) syncWorkloadsReplicasByDisposable(disposable *v1.DisposablePlan, workloads []*v1.WorkLoad) {
	if disposable == nil || len(workloads) == 0{
		return
	}
	startTimeStr := disposable.StartTime
	endTimeStr := disposable.EndTime
	var (
		err error
		startTime time.Time
		endTime time.Time
	)
	startTime, err = time.ParseInLocation("2006-01-02 15:04:05", startTimeStr, time.Local)
	if err != nil {
		return
	}
	endTime, err = time.ParseInLocation("2006-01-02 15:04:05", endTimeStr, time.Local)
	if err != nil {
		return
	}
	now := time.Now()
	//log.Info("Now: ", now)
	if now.After(startTime) && now.Before(endTime) {
		// sync replicas to N
		//log.Info("now.After(startTime) && now.Before(endTime)")
		p.syncReplicas(workloads)
	} else {
		// sync replicas to 0
		//log.Info("sync replicas to 0")
		//fmt.Println(workloads)
		p.setReplicasZero(workloads)
	}
}

func (p *NormalPlan) syncReplicas(workloads []*v1.WorkLoad) {
	//log.Info()
	//reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	//reqLogger.Info("Reconciling WorkLoadPlan")
	namespace := p.WorkLoadPlan.Namespace

	for _, wl := range workloads {
		workLoadType := strings.ToUpper(wl.WorkLoadType)
		key := client.ObjectKey{Namespace: namespace, Name: wl.Name}
		var err error
		switch workLoadType {
		case WORKLOADTYPE_DEPLOYMENT:
			oldDeploy := appsv1.Deployment{}
			err = K8sClient.Get(context.TODO(), key, &oldDeploy)
			if err != nil {
				continue
			}
			//ANNOTATION_LAST_REPLICAS_KEY
			if *oldDeploy.Spec.Replicas != 0 {
				continue
			}
			var size int32 = 1
			sizeStr, ok := oldDeploy.Annotations[ANNOTATION_LAST_REPLICAS_KEY]
			if ok {
				lastReplicas, err := strconv.ParseInt(sizeStr, 10, 32)
				//lastReplicas, err := strconv.Atoi(sizeStr)
				if err == nil {
					size  = int32(lastReplicas)
				}
			}
			oldDeploy.Spec.Replicas = &size
			err = K8sClient.Update(context.TODO(), &oldDeploy)
			if err != nil {
				log.Info("Failed to update deployment %s/%s", namespace, wl.Name)
			}
		case WORKLOADTYPE_STATEFULSET:
			oldDeploy := appsv1.StatefulSet{}
			err = K8sClient.Get(context.TODO(), key, &oldDeploy)
			if err != nil {
				continue
			}
			//ANNOTATION_LAST_REPLICAS_KEY
			if *oldDeploy.Spec.Replicas != 0 {
				continue
			}
			var size int32 = 1
			sizeStr, ok := oldDeploy.Annotations[ANNOTATION_LAST_REPLICAS_KEY]
			if ok {
				lastReplicas, err := strconv.ParseInt(sizeStr, 10, 32)
				if err == nil {
					size  = int32(lastReplicas)
				}
			}
			oldDeploy.Spec.Replicas = &size
			err = K8sClient.Update(context.TODO(), &oldDeploy)
			if err != nil {
				log.Info("Failed to update deployment %s/%s", namespace, wl.Name)
			}
		case WORKLOADTYPE_DAEMONSET:
			oldDeploy := appsv1.DaemonSet{}
			err = K8sClient.Get(context.TODO(), key, &oldDeploy)
			if err != nil {
				continue
			}
			nodeSelector := oldDeploy.Spec.Template.Spec.NodeSelector
			if nodeSelector == nil {
				return
			}
			delete(nodeSelector, DAEMONSET_NODE_SELECTOR_STOP)
			err = K8sClient.Update(context.TODO(), &oldDeploy)
			if err != nil {
				log.Info("Failed to update deployment %s/%s", namespace, wl.Name)
			}
		default:
		}
	}
}


func (p *NormalPlan) setReplicasZero(workloads []*v1.WorkLoad) {
	namespace := p.WorkLoadPlan.Namespace
	for _, wl := range workloads {
		workLoadType := strings.ToUpper(wl.WorkLoadType)
		key := client.ObjectKey{Namespace: namespace, Name: wl.Name}
		var err error
		switch workLoadType {
		case WORKLOADTYPE_DEPLOYMENT:
			log.Info("------------------------deployment")
			var err error
			key := client.ObjectKey{Namespace: namespace, Name: wl.Name}
			oldDeploy := appsv1.Deployment{}
			err = K8sClient.Get(context.TODO(), key, &oldDeploy)
			if err != nil {
				continue
			}
			var size int32 = 0
			oldDeploy.Spec.Replicas = &size
			err = K8sClient.Update(context.TODO(), &oldDeploy)
			if err != nil {
				//log.Info("Failed to update deployment %s/%s", namespace, wl.Name)
			}
		case WORKLOADTYPE_STATEFULSET:
			oldDeploy := appsv1.StatefulSet{}
			err = K8sClient.Get(context.TODO(), key, &oldDeploy)
			if err != nil {
				continue
			}
			//ANNOTATION_LAST_REPLICAS_KEY
			if *oldDeploy.Spec.Replicas != 0 {
				continue
			}
			var size int32 = 0
			oldDeploy.Spec.Replicas = &size
			err = K8sClient.Update(context.TODO(), &oldDeploy)
			if err != nil {
				log.Info("Failed to update deployment %s/%s", namespace, wl.Name)
			}

		case WORKLOADTYPE_DAEMONSET:
			oldDeploy := appsv1.DaemonSet{}
			err = K8sClient.Get(context.TODO(), key, &oldDeploy)
			if err != nil {
				continue
			}
			if oldDeploy.Spec.Template.Spec.NodeSelector == nil {
				oldDeploy.Spec.Template.Spec.NodeSelector = make(map[string]string)
			}
			oldDeploy.Spec.Template.Spec.NodeSelector[DAEMONSET_NODE_SELECTOR_STOP] = "true"
			err = K8sClient.Update(context.TODO(), &oldDeploy)
			if err != nil {
				log.Info("Failed to update deployment %s/%s", namespace, wl.Name)
			}
			/*nodeSelector := oldDeploy.Spec.Template.Spec.NodeSelector
			if nodeSelector == nil {
				nodeSelector = make(map[string]string)
				nodeSelector[DAEMONSET_NODE_SELECTOR_STOP] = "true"
				oldDeploy.Spec.Template.Spec.NodeSelector =nodeSelector
				err = K8sClient.Update(context.TODO(), &oldDeploy)
				if err != nil {
					log.Info("Failed to update deployment %s/%s", namespace, wl.Name)
				}
				return
			}
			nodeSelector[DAEMONSET_NODE_SELECTOR_STOP] = "true"
			err = K8sClient.Update(context.TODO(), &oldDeploy)
			if err != nil {
				log.Info("Failed to update deployment %s/%s", namespace, wl.Name)
			}*/
		default:
		}
	}
}


func (p *NormalPlan) syncWorkloadsReplicas() (err error){

	cr := *p.WorkLoadPlan
	crSpec := cr.Spec
	if crSpec.Disable == true {
		//log.Info("crSpec.Disable == true")
		return
	}
	var cron *v1.CrondPlan
	var disposable *v1.DisposablePlan
	if crSpec.Cron == nil {
		disposable = crSpec.Disposable
		//log.Info("crSpec.Cron == nil")
		if crSpec.Disposable == nil {
			//log.Info("crSpec.Disposable == nil")
		}
		p.syncWorkloadsReplicasByDisposable(disposable, crSpec.WorkLoads)
	} else {
		cron = crSpec.Cron
		//log.Info("start syncWorkloadsReplicasByCron")
		p.syncWorkloadsReplicasByCron(cron, crSpec.WorkLoads)
	}
	return
}

func (p *NormalPlan) Pause() {
	p.pause = true
}

func (p *NormalPlan) Destroy() {
	p.cancelFunc()
	p.looping = false
}



