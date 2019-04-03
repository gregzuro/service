package common

// // GetContainerList gets a list of docker containers
// func GetContainerList(docker string) []types.Container {
// 	fmt.Println("Connecting to docker:", docker)
// 	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
// 	cli, err := client.NewClient(docker, "v1.22", nil, defaultHeaders)
// 	if err != nil {
// 		panic(err)
// 	}
// 	options := types.ContainerListOptions{All: true}
// 	containers, err := cli.ContainerList(context.Background(), options)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return containers
// }
