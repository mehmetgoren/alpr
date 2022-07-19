package main

import (
	"alpr/dckr"
	"alpr/eb"
	"alpr/models"
	"alpr/reps"
	"alpr/utils"
	"encoding/json"
	"github.com/docker/docker/client"
	"github.com/go-redis/redis/v8"
	"log"
)

func removeContainers(dm *dckr.AlprDockerManager, all bool) {
	_, err := dm.RemoveContainers(all)
	if err != nil {
		log.Println("an error occurred while removing a container, err: ", err.Error())
		return
	}
}

func test(acs *dckr.AlprContainerScheduler, counter *utils.Counter) error {
	resp, _ := acs.Detect(counter, "4.jpg")
	jo, err := json.Marshal(resp)
	print(jo)
	return err
}

func setUpService(client *redis.Client, config *models.Config) {
	serviceName := "automatic_license_plate_recognition"
	var heartbeatRepository = reps.HeartbeatRepository{Client: client, TimeSecond: int64(config.General.HeartbeatInterval), ServiceName: serviceName}

	go heartbeatRepository.Start()
	serviceRepository := reps.ServiceRepository{Client: client}
	go func() {
		_, err := serviceRepository.Add(serviceName)
		if err != nil {
			log.Println("An error occurred while registering process id, error is:" + err.Error())
		}
	}()
}

func main() {
	defer utils.HandlePanic()
	utils.RemovePrevTempImageFiles()

	clnt, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Println("docker client couldn't be created, err: " + err.Error())
		return
	}
	dm := &dckr.AlprDockerManager{Client: clnt}
	defer func() {
		err := clnt.Close()
		if err != nil {
			log.Println("an error occurred during the closing docker client, err: ", err.Error())
			return
		}
		removeContainers(dm, true)
		utils.RemovePrevTempImageFiles()
	}()
	removeContainers(dm, true)

	err = dm.InitImage()
	if err != nil {
		return
	}

	rb := reps.RepoBucket{}
	rb.Init()

	var counter utils.Counter = 0
	redisClient := rb.GetMainConnection()
	var configRep = reps.ConfigRepository{Connection: redisClient}
	config, _ := configRep.GetConfig()
	acs := dckr.AlprContainerScheduler{Mngr: dm, Config: config}
	acs.InitContainers()

	err = test(&acs, &counter)
	if err != nil {
		log.Println("testing is not ok, exiting now...")
		return
	} else {
		log.Println("testing was successful")
	}

	setUpService(redisClient, config)

	event := eb.EventBus{Rb: &rb, Channel: "read_service"}
	eventHandler := &eb.FFmpegReaderResponseEvent{Scheduler: acs, Counter: &counter, Rb: &rb}
	err = event.Subscribe(eventHandler)
	if err != nil {
		log.Println("an error occurred while listening read service event, the process is now exiting err: ", err.Error())
	}
}
