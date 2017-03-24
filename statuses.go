package main

import "github.com/3Blades/go-sdk/client/projects"

// SetStatus writes server status in redis cache
func SetStatus(args *Args, status string) {
	cli := APIClient(args.ApiRoot, args.ApiKey)
	params := projects.NewProjectsServersPartialUpdateParams()
	params.SetNamespace(args.Namespace)
	params.SetProjectPk(args.ProjectID)
	params.SetID(args.ServerID)
	params.SetData(projects.ProjectsServersPartialUpdateBody{
		Status: status,
	})
	_, err := cli.Projects.ProjectsServersPartialUpdate(params)
	if err != nil {
		logger.Println("[SetStatus]", err)
	}
}
