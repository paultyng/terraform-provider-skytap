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

func flattenDisks(disks []skytap.Disk) *schema.Set {
	results := make([]interface{}, 0)

	for _, v := range disks {
		// ignore os disk for now
		if "0" != *v.LUN {
			results = append(results, flattenDisk(v))
		}
	}

	return schema.NewSet(diskHash, results)
}

func flattenDisk(v skytap.Disk) map[string]interface{} {
	result := make(map[string]interface{})
	size := *v.Size
	result["id"] = *v.ID
	result["size"] = size
	result["type"] = *v.Type
	result["controller"] = *v.Controller
	result["lun"] = *v.LUN
	result["name"] = *v.Name
	return result
}
