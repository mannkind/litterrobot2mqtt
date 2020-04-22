package lib

import (
	"fmt"
	"net"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

const (
	apiURL             = "https://muvnkjeut7.execute-api.us-east-1.amazonaws.com/staging"
	loginURL           = apiURL + "/login"
	statusURL          = apiURL + "/users/%s/litter-robots"
	cmdURL             = apiURL + "/users/%s/litter-robots/%s/dispatch-commands"
	lrRemoteAPIUDPPort = 2001
	lrLocalAPIUDPPort  = 2000
	onFlag             = "1"
	offFlag            = "0"
	powerCmd           = "<P"
	cycleCmd           = "<C"
	nightLightCmd      = "<N"
	panelLockCmd       = "<L"
	waitCmd            = "<W"
)

// Client - The Client for accessing data
type Client struct {
	Opts
	stateChan chan State
	robots    map[string]string
}

// NewClient - Create a new Client for a given address
func NewClient(opts Opts, stateChan chan State) *Client {
	return &Client{
		Opts:      opts,
		stateChan: stateChan,
		robots:    map[string]string{},
	}
}

// States - Fetch states from the Litter Robot API
func (c *Client) States() ([]State, error) {
	results, err := c.fetchStatesFromAPI()
	if err != nil {
		return nil, err
	}

	robots := make([]State, 0)
	for _, result := range results {
		robots = append(robots, State{
			LitterRobotID:             result.LitterRobotID,
			LitterRobotSerial:         result.LitterRobotSerial,
			NameOrIP:                  result.LitterRobotNickname,
			PowerStatus:               result.PowerStatus,
			UnitStatus:                result.UnitStatus,
			CycleCount:                result.CycleCount,
			CycleCapacity:             fmt.Sprintf("%v", result.CycleCapacity),
			CyclesAfterDrawerFull:     result.CyclesAfterDrawerFull,
			DFICycleCount:             result.DFICycleCount,
			CleanCycleWaitTimeMinutes: result.CleanCycleWaitTimeMinutes,
			PanelLockActive:           result.PanelLockActive == onFlag,
			NightLightActive:          result.NightLightActive == onFlag,
			SleepModeActive:           result.SleepModeActive == onFlag,
			DFITriggered:              result.IsDFITriggered == onFlag,
		})
	}

	return robots, nil
}

// Publish - Fetch & sendState states from the Litter Robot API
func (c *Client) Publish() {
	states, err := c.States()
	if err != nil {
		return
	}

	for _, state := range states {
		c.stateChan <- state
	}
}

// Run - Publish states via the API or incoming UDP (depending on configuration)
func (c *Client) Run() {
	if c.Local {
		go c.publishStateUDP()
	} else if c.APILookupInterval != 0 {
		sched := cron.New()
		sched.AddFunc(fmt.Sprintf("@every %s", c.APILookupInterval), c.publishStateAPI)
		sched.Start()
	}
}

func (c *Client) loginAPI() error {
	log.Debug("Trying to login to the Litter Robot API")

	if c.token != "" {
		log.Debug("Login to the Litter Robot API was skipped; token exists")
		return nil
	}

	client := resty.New()
	resp, err := client.R().
		SetResult(apiLoginResponse{}).
		SetHeader("Content-Type", "application/json").
		SetHeader("x-api-key", c.APIKey).
		SetBody(fmt.Sprintf(`{"email": "%s", "oneSignalPlayerId": "0", "password": "%s"}`, c.Email, c.Password)).
		Post(loginURL)

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Login to the Litter Robot API failed")
		return err
	}

	log.Debug("Login to the Litter Robot API succeeded")

	result := (*resp.Result().(*apiLoginResponse))
	c.userID = result.User.UserID
	c.token = result.Token

	return nil

}

func (c *Client) fetchStatesFromAPI() ([]apiRobotResponse, error) {
	log.Debug("Trying to fetch data from the Litter Robot API")

	if c.token == "" {
		if err := c.loginAPI(); err != nil {
			return nil, err
		}
	}

	client := resty.New()
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("x-api-key", c.APIKey).
		SetHeader("Authorization", c.token).
		SetResult([]apiRobotResponse{}).
		Get(fmt.Sprintf(statusURL, c.userID))

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Unable to fetch robot information")
		return nil, err
	}

	results := (*resp.Result().(*[]apiRobotResponse))

	log.Debug("Fetching data from the Litter Robot API succeeded")
	return results, nil
}

func (c *Client) publishStateAPI() {
	c.Publish()
}

func (c *Client) publishStateUDP() {
	for {
		conn, err := net.ListenUDP("udp", &net.UDPAddr{
			Port: lrRemoteAPIUDPPort,
		})

		if err != nil {
			log.WithFields(log.Fields{
				"address": conn.LocalAddr().String(),
			}).Error("Unable to start UDP Server")

			continue
		}

		defer conn.Close()

		log.WithFields(log.Fields{
			"address": conn.LocalAddr().String(),
		}).Info("Listening for LR UDP Communication")

		for {
			// Read the payload
			message := make([]byte, 64)
			rlen, remote, err := conn.ReadFromUDP(message[:])
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Info("Error reading UDP packet")
			}

			// Immediately relay the payload to the Litter Robot API
			dispatch, err := net.ResolveUDPAddr("udp", "dispatch.iothings.site:"+fmt.Sprintf("%d", lrRemoteAPIUDPPort))
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Info("Error resolving dispatch.iothings.site")
			}

			sendConn, err := net.DialUDP("udp", nil, dispatch)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("Error determining UDPAddr")
			}

			_, err = sendConn.Write(message)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err,
				}).Error("Error sending UDP packet")
			}

			data := strings.TrimSpace(string(message[:rlen]))

			log.WithFields(log.Fields{
				"ip":   remote,
				"data": data,
			}).Info("Received UDP Packet")

			parts := strings.Split(data, ",")

			dfsCount := ""
			dfs := strings.HasPrefix(parts[4], "DFS")
			if dfs {
				dfsCount = strings.Replace(parts[4], "DFS", "", 1)
				if dfsCount == "" {
					dfsCount = "0"
				}
			}
			state := State{
				LitterRobotID:             parts[1],
				NameOrIP:                  remote.IP.String(),
				PowerStatus:               parts[3],
				UnitStatus:                parts[4],
				CyclesAfterDrawerFull:     dfsCount,
				CleanCycleWaitTimeMinutes: strings.Replace(parts[5], "W", "", 1),
				PanelLockActive:           strings.Replace(parts[8], "PL", "", 1) == onFlag,
				NightLightActive:          strings.Replace(parts[6], "NL", "", 1) == onFlag,
				SleepModeActive:           strings.Replace(strings.Replace(parts[7], "SM0", "0", 1), "SM", "", 1) == onFlag,
				DFITriggered:              dfs,
			}

			c.stateChan <- state
		}
	}
}

func (c *Client) sendCommand(identifier string, ip string, command string) error {
	if !c.Local {
		return c.sendAPICommand(identifier, command)
	}

	return c.sendUDPCommand(identifier, ip, command)
}

func (c *Client) sendUDPCommand(identifier string, ip string, command string) error {
	sendConn, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: lrLocalAPIUDPPort,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error determining UDPAddr")

		return err
	}

	model := "LR3"    // @TODO
	counter := "06EB" // @TODO
	crc := "7AE2E42F" // @TOOD
	fullCommand := command + "," + model + "," + identifier + "," + counter
	_, err = sendConn.Write([]byte(fullCommand + "," + crc))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error sending UDP packet")

		return err
	}

	return nil
}

func (c *Client) sendAPICommand(identifier string, command string) error {
	// Fetch a token if we don't have one
	if c.token == "" {
		if err := c.loginAPI(); err != nil {
			return err
		}
	}

	log.WithFields(log.Fields{
		"command":    command,
		"identifier": identifier,
	}).Info("Sending command to Litter Robot")

	client := resty.New()
	_, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("x-api-key", c.APIKey).
		SetHeader("Authorization", c.token).
		SetBody(fmt.Sprintf(`{"command": "%s", "litterRobotId": "%s"}`, command, identifier)).
		Post(fmt.Sprintf(cmdURL, c.userID, identifier))

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error sending command to Litter Robot")
		return err
	}

	return nil
}

// PowerOn - Power Command
func (c *Client) PowerOn(identifier string, nameOrIP string) {
	log.Debug("Sending Power On Command")
	c.sendCommand(identifier, nameOrIP, powerCmd+onFlag)
}

// PowerOff - Power Command
func (c *Client) PowerOff(identifier string, nameOrIP string) {
	log.Debug("Sending Power Off Command")
	c.sendCommand(identifier, nameOrIP, powerCmd+offFlag)
}

// NightLightOn - NightLight Command
func (c *Client) NightLightOn(identifier string, nameOrIP string) {
	log.Debug("Sending Night Light On Command")
	c.sendCommand(identifier, nameOrIP, nightLightCmd+onFlag)
}

// NightLightOff - NightLight Command
func (c *Client) NightLightOff(identifier string, nameOrIP string) {
	log.Debug("Sending Night Light Off Command")
	c.sendCommand(identifier, nameOrIP, nightLightCmd+offFlag)
}

// PanelLockOn - PanelLock Command
func (c *Client) PanelLockOn(identifier string, nameOrIP string) {
	log.Debug("Sending Panel Lock On Command")
	c.sendCommand(identifier, nameOrIP, panelLockCmd+onFlag)
}

// PanelLockOff - PanelLock Command
func (c *Client) PanelLockOff(identifier string, nameOrIP string) {
	log.Debug("Sending Panel Lock Off Command")
	c.sendCommand(identifier, nameOrIP, panelLockCmd+offFlag)
}

// Cycle - Cycle Command
func (c *Client) Cycle(identifier string, nameOrIP string) {
	log.Debug("Sending Cycle Command")
	c.sendCommand(identifier, nameOrIP, cycleCmd)
}

// Wait - Wait Command
func (c *Client) Wait(identifier string, nameOrIP string, val string) {
	log.Debug("Sending Wait Command")
	c.sendCommand(identifier, nameOrIP, waitCmd+val)
}
