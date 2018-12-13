package skytap

import (
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/skytap/skytap-sdk-go/skytap"
	"github.com/stretchr/testify/assert"
)

const exampleInterfaceListResponse = `[
    {
        "id": "nic-20246343-38367563-0",
        "ip": "192.168.0.1",
        "hostname": "wins2016s",
        "mac": "00:50:56:11:7D:D9",
        "services_count": 0,
        "services": [],
        "public_ips_count": 0,
        "public_ips": [],
        "vm_id": "37527239",
        "vm_name": "Windows Server 2016 Standard",
        "status": "Running",
        "network_id": "23917287",
        "network_name": "tftest-network-1",
        "network_url": "https://cloud.skytap.com/v2/configurations/40064014/networks/23917287",
        "network_type": "automatic",
        "network_subnet": "192.168.0.0/16",
        "nic_type": "vmxnet3",
        "secondary_ips": [],
        "public_ip_attachments": []
    },
    {
        "id": "nic-20246343-38367563-5",
        "ip": "192.168.0.2",
        "hostname": "wins2016s2",
        "mac": "00:50:56:07:40:3F",
        "services_count": 0,
        "services": [],
        "public_ips_count": 0,
        "public_ips": [],
        "vm_id": "37527239",
        "vm_name": "Windows Server 2016 Standard",
        "status": "Running",
        "network_id": "23917287",
        "nic_type": "e1000",
        "secondary_ips": [],
        "public_ip_attachments": []
    }
]`

const examplePublishedServiceListResponse = `[
    {
        "id": "8080",
        "internal_port": 8080,
        "external_ip": "services-uswest.skytap.com",
        "external_port": 26160
    },
    {
        "id": "8081",
        "internal_port": 8081,
        "external_ip": "services-useast.skytap.com",
        "external_port": 17785
    }
]`

const exampleInterfaceAndPortsListResponse = `[
    {
        "id": "nic-20246343-38367563-0",
        "ip": "192.168.0.1",
        "hostname": "wins2016s",
        "mac": "00:50:56:11:7D:D9",
        "services_count": 0,
        "services": [
	    {
            "id": "8080",
            "internal_port": 8080,
            "external_ip": "services-uswest.skytap.com",
            "external_port": 26160
        },
        {
            "id": "8081",
            "internal_port": 8081,
            "external_ip": "services-useast.skytap.com",
            "external_port": 17785
    	}
        ],
        "public_ips_count": 0,
        "public_ips": [],
        "vm_id": "37527239",
        "vm_name": "Windows Server 2016 Standard",
        "status": "Running",
        "network_id": "23917287",
        "network_name": "tftest-network-1",
        "network_url": "https://cloud.skytap.com/v2/configurations/40064014/networks/23917287",
        "network_type": "automatic",
        "network_subnet": "192.168.0.0/16",
        "nic_type": "vmxnet3",
        "secondary_ips": [],
        "public_ip_attachments": []
    },
    {
        "id": "nic-20246343-38367563-5",
        "ip": "192.168.0.2",
        "hostname": "wins2016s2",
        "mac": "00:50:56:07:40:3F",
        "services_count": 0,
        "services": [
	    {
            "id": "8080",
            "internal_port": 8080,
            "external_ip": "services-uswest.skytap.com",
            "external_port": 26160
        },
        {
            "id": "8081",
            "internal_port": 8081,
            "external_ip": "services-useast.skytap.com",
            "external_port": 17785
    	}
        ],
        "public_ips_count": 0,
        "public_ips": [],
        "vm_id": "37527239",
        "vm_name": "Windows Server 2016 Standard",
        "status": "Running",
        "network_id": "23917287",
        "nic_type": "e1000",
        "secondary_ips": [],
        "public_ip_attachments": []
    }
]`

func TestFlattenInterfaces(t *testing.T) {

	expected := make(map[string][]string)
	expected["ip"] = []string{"192.168.0.1", "192.168.0.2"}
	expected["hostname"] = []string{"wins2016s", "wins2016s2"}

	var ipMaps = []map[string]interface{}{make(map[string]interface{}), make(map[string]interface{})}

	var interfaces []skytap.Interface
	json.Unmarshal([]byte(exampleInterfaceListResponse), &interfaces)
	var networkInterfaces = make([]map[string]interface{}, 0)
	for _, v := range interfaces {
		networkInterfaces = append(networkInterfaces, flattenNetworkInterface(v, ipMaps))
	}
	assert.True(t, len(networkInterfaces) == 2, fmt.Sprintf("expecting: %d but received: %d", 2, len(networkInterfaces)))
	for i := 0; i < len(networkInterfaces); i++ {
		networkInterface := networkInterfaces[i]
		for key, value := range expected {
			assert.Equal(t, value[i], networkInterface[key],
				fmt.Sprintf("expecting: %s but received: %s", value[i], networkInterface[key]))
		}
	}
}

func TestFlattenPublishedServices(t *testing.T) {

	expected := make(map[string][]string)
	expected["id"] = []string{"8080", "8081"}
	expected["internal_port"] = []string{"8080", "8081"}
	expected["external_ip"] = []string{"services-uswest.skytap.com", "services-useast.skytap.com"}
	expected["external_port"] = []string{"26160", "17785"}

	var ipMaps = []map[string]interface{}{make(map[string]interface{}), make(map[string]interface{})}

	var services []skytap.PublishedService
	json.Unmarshal([]byte(examplePublishedServiceListResponse), &services)
	var publishedServices = make([]map[string]interface{}, 0)
	for _, v := range services {
		publishedServices = append(publishedServices, flattenPublishedService(v, "", ipMaps))
	}
	assert.True(t, len(publishedServices) == 2, fmt.Sprintf("expecting: %d but received: %d", 2, len(publishedServices)))
	for i := 0; i < len(publishedServices); i++ {
		publishedService := publishedServices[i]
		for key, value := range expected {
			if _, ok := publishedService[key].(string); ok {
				assert.Equal(t, value[i], publishedService[key].(string),
					fmt.Sprintf("expecting: %s but received: %s", value[i], publishedService[key]))
			} else {
				valueAsString := strconv.Itoa(publishedService[key].(int))
				assert.Equal(t, value[i], valueAsString,
					fmt.Sprintf("expecting: %s but received: %s", value[i], valueAsString))
			}
		}
	}
}

func TestFlattenInterfacesIPMap(t *testing.T) {
	expectedKeys1 := []string{"192-168-0-1_8080", "192-168-0-1_8081", "192-168-0-2_8080", "192-168-0-2_8081"}
	expectedValues1 := []string{"services-uswest.skytap.com", "services-useast.skytap.com", "services-uswest.skytap.com", "services-useast.skytap.com"}
	expectedKeys2 := []string{"192-168-0-1_8080", "192-168-0-1_8081", "192-168-0-2_8080", "192-168-0-2_8081"}
	expectedValues2 := []string{"26160", "17785", "26160", "17785"}

	var ipMaps = []map[string]interface{}{make(map[string]interface{}), make(map[string]interface{})}

	var interfaces []skytap.Interface
	json.Unmarshal([]byte(exampleInterfaceAndPortsListResponse), &interfaces)
	var networkInterfaces = make([]map[string]interface{}, 0)
	for _, v := range interfaces {
		networkInterfaces = append(networkInterfaces, flattenNetworkInterface(v, ipMaps))
	}

	count := 0
	for _, key := range expectedKeys1 {
		assert.Equal(t, expectedValues1[count], ipMaps[0][key], "value")
		count++
	}
	count = 0
	for _, key := range expectedKeys2 {
		assert.Equal(t, expectedValues2[count], ipMaps[1][key], "value")
		count++
	}
}
