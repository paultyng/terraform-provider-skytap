package skytap

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/skytap/skytap-sdk-go/skytap"
)

func flattenNetworkInterfaces(interfaces []skytap.Interface) *schema.Set {
	results := make([]interface{}, 0)

	for _, v := range interfaces {
		results = append(results, flattenNetworkInterface(v))
	}

	return schema.NewSet(networkInterfaceHash, results)
}

func flattenNetworkInterface(v skytap.Interface) map[string]interface{} {
	result := make(map[string]interface{})
	result["interface_type"] = string(*v.NICType)
	result["network_id"] = *v.NetworkID
	result["ip"] = *v.IP
	result["hostname"] = *v.Hostname
	if len(v.Services) > 0 {
		result["published_service"] = flattenPublishedServices(v.Services)
	}
	return result
}

func flattenPublishedServices(publishedServices []skytap.PublishedService) *schema.Set {
	results := make([]interface{}, 0)

	for _, v := range publishedServices {
		results = append(results, flattenPublishedService(v))
	}

	return schema.NewSet(publishedServiceHash, results)
}

func flattenPublishedService(v skytap.PublishedService) map[string]interface{} {
	result := make(map[string]interface{})
	result["id"] = *v.ID
	result["internal_port"] = *v.InternalPort
	result["external_ip"] = *v.ExternalIP
	result["external_port"] = *v.ExternalPort
	return result
}

func flattenDisks(disks []skytap.Disk, diskResource *schema.Set) *schema.Set {
	results := make([]interface{}, 0)

	firstTime := make(map[int][]string)
	otherTimes := make(map[string]string)
	if diskResource != nil {
		for _, v := range diskResource.List() {
			values := v.(map[string]interface{})
			id := values["id"].(string)
			name := values["name"].(string)
			if id == "" {
				size := values["size"].(int)
				if len(firstTime[size]) == 0 {
					firstTime[size] = make([]string, 0)
				}
				firstTime[size] = append(firstTime[size], name)
			} else {
				otherTimes[id] = name
			}
		}
	}
	for _, v := range disks {
		// ignore os disk for now
		if "0" != *v.LUN {
			results = append(results, flattenDisk(v, firstTime, otherTimes))
		}
	}

	return schema.NewSet(diskHash, results)
}

func flattenDisk(v skytap.Disk, firstTime map[int][]string, otherTimes map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	size := *v.Size
	result["id"] = *v.ID
	result["size"] = size
	result["type"] = *v.Type
	result["controller"] = *v.Controller
	result["lun"] = *v.LUN
	if len(otherTimes) > 0 {
		result["name"] = otherTimes[*v.ID]
	} else if len(firstTime[size]) > 0 {
		result["name"] = firstTime[size][0]
		firstTime[size] = append(firstTime[size][1:])
	}
	return result
}
