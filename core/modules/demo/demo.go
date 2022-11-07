package demo

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/hamster-shared/hamster-provider/log"
)

type DApp struct {
	PeerID      string
	CPU         uint8
	Memory      uint8
	StartMethod uint8  // 启动方式，1 表示 docker,2 表示 docker compose
	Command     string // 启动命令，如果使用 docker，就作为 docker 的命令，如果使用 docker compose，表示这是一个 ipfs cid，里面存放着 compose 文件，需要先去 ipfs 下载

	name string
}

func NewDApp(peerID string, cpu, mem uint8, method uint8, command string) *DApp {
	name := uuid.New().String()
	return &DApp{
		name:        name,
		PeerID:      peerID,
		CPU:         cpu,
		Memory:      mem,
		StartMethod: method,
		Command:     command,
	}
}

func (d *DApp) Start() error {
	switch d.StartMethod {
	case 1:
		return d.startDocker()
	case 2:
		return d.startCompose()
	default:
		return fmt.Errorf("unknown start method: %d", d.StartMethod)
	}
}

func (d *DApp) Stop() error {
	switch d.StartMethod {
	case 1:
		return d.stopDocker()
	case 2:
		return d.stopCompose()
	default:
		return fmt.Errorf("unknown start method: %d", d.StartMethod)
	}
}

func (d *DApp) Restart() error {
	switch d.StartMethod {
	case 1:
		return d.restartDocker()
	case 2:
		return d.restartCompose()
	default:
		return fmt.Errorf("unknown start method: %d", d.StartMethod)
	}
}

func (d *DApp) Status() error {
	switch d.StartMethod {
	case 1:
		return d.statusDocker()
	case 2:
		return d.statusCompose()
	default:
		return fmt.Errorf("unknown start method: %d", d.StartMethod)
	}
}

func (d *DApp) Update(new DApp) error {
	switch d.StartMethod {
	case 1:
		return d.updateDocker(new)
	case 2:
		return d.updateCompose(new)
	default:
		return fmt.Errorf("unknown start method: %d", d.StartMethod)
	}
}

func (d *DApp) startDocker() error {
	cmdSlice := []string{"run"}
	cmdSlice = append(cmdSlice, "--name", d.name)
	cmdSlice = append(cmdSlice, strings.Split(d.Command, " ")...)

	cmd := exec.Command("docker", cmdSlice...)
	log.GetLogger().Debugf("docker container start with command: %s", cmd.String())

	out, err := cmd.Output()
	log.GetLogger().Debugf("command output: %s", string(out))
	if err != nil {
		log.GetLogger().Errorf("docker container failed, error: %s", err)
		return err
	}
	log.GetLogger().Infof("docker container success, command: %s", cmd.String())

	return nil
}

func (d *DApp) startCompose() error {
	// 连接 ipfs，将 Command 字段视为 cid, 获取 docker compose
	cid := d.Command
	ipfsShell := IPFS{}

	ipfsShell.Connect("http://localhost:5001")
	if err := ipfsShell.Get(cid, "./tmp"); err != nil {
		return err
	}
	log.GetLogger().Debugf("get file from ipfs with cid %s success", cid)

	cmd := exec.Command("docker", "compose", "-f", filepath.Join("./tmp", cid), "up", "-d")
	if err := cmd.Run(); err != nil {
		return err
	}
	log.GetLogger().Infof("start docker compose with cid %s success, command: %s")

	return nil
}

func (d *DApp) stopDocker() error {
	cmdSlice := []string{"stop"}
	cmdSlice = append(cmdSlice, d.name)

	cmd := exec.Command("docker", cmdSlice...)
	log.GetLogger().Debugf("docker container stop with command: %s", cmd.String())

	out, err := cmd.Output()
	log.GetLogger().Debugf("command output: %s", string(out))
	if err != nil {
		log.GetLogger().Errorf("docker container stop failed, error: %s", err)
		return err
	}

	return nil
}

func (d *DApp) stopCompose() error {
	cid := d.Command
	ipfsShell := IPFS{}

	ipfsShell.Connect("http://localhost:5001")
	if err := ipfsShell.Get(cid, "./tmp"); err != nil {
		return err
	}
	log.GetLogger().Debugf("get file from ipfs with cid %s success", cid)

	cmd := exec.Command("docker", "compose", "-f", filepath.Join("./tmp", cid), "down")
	if err := cmd.Run(); err != nil {
		return err
	}
	log.GetLogger().Infof("down docker compose with cid %s success, command: %s")

	return nil
}

func (d *DApp) restartDocker() error {
	cmdSlice := []string{"restart"}
	cmdSlice = append(cmdSlice, d.name)

	cmd := exec.Command("docker", cmdSlice...)
	log.GetLogger().Debugf("docker container restart with command: %s", cmd.String())

	out, err := cmd.Output()
	log.GetLogger().Debugf("command output: %s", string(out))

	if err != nil {
		log.GetLogger().Errorf("docker container restart failed, error: %s", err)
		return err
	}
	log.GetLogger().Infof("docker container restart success, command: %s", cmd.String())

	return nil
}

func (d *DApp) restartCompose() error {
	cid := d.Command
	ipfsShell := IPFS{}

	ipfsShell.Connect("http://localhost:5001")
	if err := ipfsShell.Get(cid, "./tmp"); err != nil {
		return err
	}
	log.GetLogger().Debugf("get file from ipfs with cid %s success", cid)

	cmd := exec.Command("docker", "compose", "-f", filepath.Join("./tmp", cid), "restart")
	if err := cmd.Run(); err != nil {
		return err
	}
	log.GetLogger().Infof("restart docker compose with cid %s success, command: %s")

	return nil
}

func (d *DApp) statusDocker() error {
	return nil
}

func (d *DApp) statusCompose() error {
	return nil
}

func (d *DApp) updateDocker(new DApp) error {
	cmdSlice := []string{"update"}
	cmdSlice = append(cmdSlice, strings.Split(new.Command, " ")...)

	cmd := exec.Command("docker", cmdSlice...)
	log.GetLogger().Debugf("docker container update with command: %s", cmd.String())

	out, err := cmd.Output()
	log.GetLogger().Debugf("command output: %s", string(out))
	if err != nil {
		log.GetLogger().Errorf("docker container update failed, error: %s", err)
		return err
	}
	log.GetLogger().Infof("docker container update success, command: %s", cmd.String())

	return nil
}

func (d *DApp) updateCompose(new DApp) error {
	cid := new.Command
	ipfsShell := IPFS{}

	ipfsShell.Connect("http://localhost:5001")
	if err := ipfsShell.Get(cid, "./tmp"); err != nil {
		return err
	}
	log.GetLogger().Debugf("get file from ipfs with cid %s success", cid)

	cmd := exec.Command("docker", "compose", "-f", filepath.Join("./tmp", cid), "pull")
	if err := cmd.Run(); err != nil {
		return err
	}
	log.GetLogger().Infof("command executed success: %s")

	cmd = exec.Command("docker", "compose", "-f", filepath.Join("./tmp", cid), "up", "-d")
	if err := cmd.Run(); err != nil {
		return err
	}
	log.GetLogger().Infof("update docker compose with cid %s success, command: %s")

	return nil
}
