package models

type PositionTracker struct {
	timeFirstMintEver int
}

func (p *PositionTracker) updateMint(time int) {
	if p.timeFirstMintEver == 0 {
		p.timeFirstMintEver = time
	}
}

func (p *PositionTracker) updateBurn(time int) {
}

func (p *PositionTracker) updateHarvest(time int) {

}

func updatePosition(msg posUpdateMsg) {
	if msg.update == posMint {
		msg.pos.updateMint(msg.time)
	} else if msg.update == posBurn {
		msg.pos.updateBurn(msg.time)
	} else if msg.update == posHarvest {
		msg.pos.updateHarvest(msg.time)
	}
}

const UPDATE_CHANNEL_SIZE = 16000

func watchPositionUpdates() chan posUpdateMsg {
	sink := make(chan posUpdateMsg, UPDATE_CHANNEL_SIZE)
	go func() {
		for true {
			updatePosition(<-sink)
		}
	}()
	return sink
}

type posUpdateMsg struct {
	pos    *PositionTracker
	update posUpdateType
	time   int
}

type posUpdateType int

const (
	posMint posUpdateType = iota
	posBurn
	posHarvest
)
