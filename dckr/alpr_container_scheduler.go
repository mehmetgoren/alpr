package dckr

import (
	"alpr/models"
	"alpr/utils"
	"github.com/docker/docker/api/types"
	"log"
	"strconv"
)

type AlprContainerScheduler struct {
	Mngr    *AlprDockerManager
	Config  *models.Config
	ctrList []*types.Container
}

var containerNamePrefix = "alpr"

func (s *AlprContainerScheduler) InitContainers() {
	dm := s.Mngr
	instanceCount := s.Config.Ai.PlateRecogInstanceCount

	s.ctrList = make([]*types.Container, 0)
	for j := 0; j < instanceCount; j++ {
		name := containerNamePrefix + strconv.Itoa(j)
		ctr, _ := dm.GetContainer(name)
		if ctr == nil {
			ctr, _ = dm.StartContainer(name)
		}
		s.ctrList = append(s.ctrList, ctr)
	}
}

func (s *AlprContainerScheduler) Detect(counter *utils.Counter, fileName string) (*models.AlprResult, error) {
	dm := s.Mngr
	index := int(counter.Inc()) % s.Config.Ai.PlateRecogInstanceCount
	containerName := containerNamePrefix + strconv.Itoa(index)
	model, err := dm.ExecRun(s.ctrList[index], fileName)
	if err != nil {
		log.Println("an error occurred on exec_run, err: ", err.Error())
		return nil, err
	}
	if len(model.Results) > 0 {
		result := model.Results[0]
		log.Println("("+containerName, ") :", result.Plate, "-", result.Confidence)
	} else {
		log.Println("("+containerName+") no result found for:", fileName)
	}

	return model, nil
}
