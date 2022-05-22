package discover

import (
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/golang/glog"
	kubeCore "p9t.io/kuberboat/pkg/api/core"
	"p9t.io/skafos/pkg/api/core"
	"p9t.io/skafos/pkg/skpilot/buffer"
	"p9t.io/skafos/pkg/skpilot/client"
	"p9t.io/skafos/pkg/skpilot/component"
	"p9t.io/skafos/pkg/skpilot/util"
)

// Discoverer does pod discovery and service discovery at set intervals. If there are any changes,
// it will generate new rules and detect proxy updates based on the changes, and then write them
// into the buffers.
type Discoverer struct {
	// kubeClient queries Kuberboat on the information of pods and services.
	kubeClient *client.KubeClient
	// components stores metadata of all pods, services and rules.
	components *component.SkComponents
	// ruleBuffer is the buffer where rules to update are stored.
	ruleBuffer *buffer.RuleBuffer
	// proxyBuffer is the buffer where proxies to create are stored.
	proxyBuffer *buffer.ProxyBuffer
}

func NewDiscoverer(
	kubeEndpoint string,
	components *component.SkComponents,
	ruleBuffer *buffer.RuleBuffer,
	proxyBuffer *buffer.ProxyBuffer,
) *Discoverer {
	kubeClient, err := client.NewKubeClient(kubeEndpoint, core.KUBE_PORT)
	if err != nil {
		glog.Fatal(err)
	}
	return &Discoverer{
		kubeClient:  kubeClient,
		components:  components,
		ruleBuffer:  ruleBuffer,
		proxyBuffer: proxyBuffer,
	}
}

// DoDiscovering discovers pods and services of Kuberboat at set intervals.
func (d *Discoverer) DoDiscovering(discoverInterval time.Duration) {
	for range time.Tick(discoverInterval) {
		pods, err := d.kubeClient.GetAllPods()
		if err != nil {
			glog.Error(err)
			continue
		}
		services, servicePods, err := d.kubeClient.GetAllServices()
		if err != nil {
			glog.Error(err)
			continue
		}
		// We regard the pods and services here together as one snapshot.
		d.updatePodsAndServices(pods, services, servicePods)
	}
}

// updatePodsAndServices updates pods and services based on the discovering result. It also
// write new rules and proxy updates into the buffers.
func (d *Discoverer) updatePodsAndServices(
	pods []*kubeCore.Pod,
	services []*kubeCore.Service,
	servicePods [][]string,
) {
	d.components.Mtx.Lock()
	defer d.components.Mtx.Unlock()

	// Check pods
	newSandboxInfos, currentPods := d.checkPods(pods)

	// Check services
	servicesToUpdateRule,
		currentServices,
		currentServiceToPods,
		deletedServiceRuleTypes := d.checkServices(services, servicePods, currentPods)

	// Update metadata
	d.components.Pods = currentPods
	d.components.Services = currentServices
	d.components.ServicesToPods = currentServiceToPods

	// Update proxies
	func() {
		d.proxyBuffer.LockBuffer()
		defer d.proxyBuffer.UnlockBuffer()
		for _, info := range newSandboxInfos {
			d.proxyBuffer.SetSandboxInfo(info)
		}
	}()

	// Update rules
	func() {
		d.ruleBuffer.LockBuffer()
		defer d.ruleBuffer.UnlockBuffer()
		for _, serviceName := range servicesToUpdateRule {
			ruleMeta, ok := d.components.ServiceToRule[serviceName]
			if !ok {
				oldRuleMeta, ok := deletedServiceRuleTypes[serviceName]
				if !ok {
					continue
				}
				// Remove the rule applied to the deleted service.
				switch oldRuleMeta.Kind {
				case core.RatioType:
					d.ruleBuffer.SetRatioRule(oldRuleMeta.Name, nil)
				case core.RegexType:
					d.ruleBuffer.SetRegexRule(oldRuleMeta.Name, nil)
				}
			} else {
				// Apply change of the rule.
				switch ruleMeta.Kind {
				case core.RatioType:
					rule, ok := d.components.RatioRules[ruleMeta.Name]
					if !ok {
						panic(fmt.Sprintf(
							"expect to have ratio rule %s for service %s",
							ruleMeta.Name,
							serviceName,
						))
					}
					service, pods, err := d.components.GetServiceAndServicePods(serviceName)
					if err != nil {
						panic(fmt.Sprintf("expect service %s", serviceName))
					}
					ruleGenerator := util.GenerateRatioRule(rule, service, pods)
					d.ruleBuffer.SetRatioRule(ruleMeta.Name, ruleGenerator)
				case core.RegexType:
					rule, ok := d.components.RegexRules[ruleMeta.Name]
					if !ok {
						panic(fmt.Sprintf(
							"expect to have regex rule %s for service %s",
							ruleMeta.Name,
							serviceName,
						))
					}
					service, pods, err := d.components.GetServiceAndServicePods(serviceName)
					if err != nil {
						panic(fmt.Sprintf("expect service %s", serviceName))
					}
					ruleGenerator := util.GenerateRegexRule(rule, service, pods)
					d.ruleBuffer.SetRegexRule(ruleMeta.Name, ruleGenerator)
				}
			}
		}
	}()
}

// checkPods checks whether there are updates on pods and generates corresponding new proxy information.
// It will also generate a new snapshot of current pods whether there are updates or not.
func (d *Discoverer) checkPods(pods []*kubeCore.Pod) ([]*core.SandboxInfo, map[string]*kubeCore.Pod) {
	newSandboxInfos := make([]*core.SandboxInfo, 0)
	currentPods := make(map[string]*kubeCore.Pod)
	for _, pod := range pods {
		// We only consider ready pods
		if pod.Status.Phase != kubeCore.PodReady {
			continue
		}
		existentPod, ok := d.components.Pods[pod.Name]
		if !ok || !reflect.DeepEqual(*existentPod, *pod) {
			// The pod is a new pod
			sandboxName := core.GetPodSpecificPauseName(pod)
			newSandboxInfos = append(newSandboxInfos, &core.SandboxInfo{
				SandboxName: sandboxName,
				SandboxIP:   pod.Status.PodIP,
				HostIP:      pod.Status.HostIP,
			})
		}
		currentPods[pod.Name] = pod
	}
	return newSandboxInfos, currentPods
}

// checkServices checks whether there are updates on services and records necessary information for generating
// new rules. It will also generate a new snapshot of current services whether there are updates or not.
func (d *Discoverer) checkServices(
	services []*kubeCore.Service,
	servicePods [][]string,
	currentPods map[string]*kubeCore.Pod,
) ([]string, map[string]*kubeCore.Service, map[string]*[]string, map[string]*core.RuleMeta) {

	servicesToUpdateRule := make([]string, 0)
	currentServices := make(map[string]*kubeCore.Service)
	currentServiceToPods := make(map[string]*[]string)
	for i, service := range services {
		existentService, ok := d.components.Services[service.Name]
		sort.Strings(servicePods[i])
		if ok {
			// If the service is different from the previous one, or pods in the service have changed,
			// then the rule applied to this service (if exists) needs to be updated.
			if !reflect.DeepEqual(*existentService, *service) ||
				d.checkServicePodsUpdate(service.Name, &servicePods[i], &currentPods) {
				servicesToUpdateRule = append(servicesToUpdateRule, service.Name)
			}
			delete(d.components.Services, service.Name)
		}
		currentServices[service.Name] = service
		currentServiceToPods[service.Name] = &servicePods[i]
	}

	// Remove rules for deleted services if exist
	deletedServiceRuleTypes := make(map[string]*core.RuleMeta)
	for serviceName := range d.components.Services {
		ruleMeta, ok := d.components.ServiceToRule[serviceName]
		if ok {
			switch ruleMeta.Kind {
			case core.RatioType:
				deletedServiceRuleTypes[serviceName] = ruleMeta
				delete(d.components.RatioRules, ruleMeta.Name)
			case core.RegexType:
				deletedServiceRuleTypes[serviceName] = ruleMeta
				delete(d.components.RegexRules, ruleMeta.Name)
			}
			delete(d.components.ServiceToRule, serviceName)
			servicesToUpdateRule = append(servicesToUpdateRule, serviceName)
		}
	}

	return servicesToUpdateRule, currentServices, currentServiceToPods, deletedServiceRuleTypes
}

// checkServicePodsUpdate checks whether the pods in a service need update.
// `newPods` is the latest discovered pods. Previous pod snapshot is stored in `components`.
func (d *Discoverer) checkServicePodsUpdate(
	serviceName string,
	servicePods *[]string,
	newPods *map[string]*kubeCore.Pod,
) bool {
	if len(*servicePods) != len(*d.components.ServicesToPods[serviceName]) {
		return true
	}
	previousServicePods := *d.components.ServicesToPods[serviceName]
	for i, podName := range *servicePods {
		previousPodName := previousServicePods[i]
		if podName != previousPodName ||
			!reflect.DeepEqual(*(*newPods)[podName], *d.components.Pods[podName]) {
			return true
		}
	}
	return false
}
