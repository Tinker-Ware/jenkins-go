package main

import (
	"github.com/ci/jenkins"
	"fmt"
)

func main(){
	auth := &jenkins.Auth {
		Username : "poz2k4444",
		ApiToken :"e4bbe813138eb41893052e4d12100408",
	}

	client := jenkins.NewJenkins(auth, "http://ci.tinkerware.io:8080")

	body := jenkins.Project { Description : "algo" }
	
	client.CreateJob(body, "prueba")

	fmt.Println(client.GetJobs())
}
