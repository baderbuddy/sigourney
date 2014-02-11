package main

import (
	"log"
	"os"

	"code.google.com/p/portaudio-go/portaudio"
)

func main() {
	portaudio.Initialize()
	defer portaudio.Terminate()

	e := NewEngine()

	oscMod := NewOsc()
	oscMod.SetInput("pitch", Value(-0.1))

	oscModAmp := NewAmp()
	oscModAmp.SetInput("car", oscMod)
	oscModAmp.SetInput("mod", Value(0.1))

	osc := NewOsc()
	osc.SetInput("pitch", oscModAmp)

	envMod := NewOsc()
	envMod.SetInput("pitch", Value(-1))

	envModAmp := NewAmp()
	envModAmp.SetInput("car", envMod)
	envModAmp.SetInput("mod", Value(0.02))

	envModSum := NewSum()
	envModSum.SetInput("car", envModAmp)
	envModSum.SetInput("mod", Value(0.021))

	env := NewEnv()
	env.SetInput("att", Value(0.0001))
	env.SetInput("dec", envModSum)

	amp := NewAmp()
	amp.SetInput("car", osc)
	amp.SetInput("mod", env)

	ampAmp := NewAmp()
	ampAmp.SetInput("car", amp)
	ampAmp.SetInput("mod", Value(0.5))

	e.SetInput("root", ampAmp)

	if err := e.Start(); err != nil {
		log.Println(err)
		return
	}

	os.Stdout.Write([]byte("Press enter to stop...\n"))
	os.Stdin.Read([]byte{0})

	if err := e.Stop(); err != nil {
		log.Println(err)
	}
}

func NewEngine() *Engine {
	e := &Engine{done: make(chan error)}
	newSink(&e.sink, "root", &e.root)
	return e
}

type Engine struct {
	sink
	root source

	done chan error
}

func (e *Engine) processAudio(_, out []int16) {
	buf := e.root.Process()
	for i := range buf {
		out[i] = int16(buf[i] * waveAmp)
	}
}

func (e *Engine) Start() error {
	stream, err := portaudio.OpenDefaultStream(0, 1, waveHz, nSamples, e.processAudio)
	if err != nil {
		return err
	}
	errc := make(chan error)
	go func() {
		err = stream.Start()
		errc <- err
		if err != nil {
			return
		}
		<-e.done
		err = stream.Stop()
		if err == nil {
			err = stream.Close()
		}
		e.done <- err
	}()
	return <-errc
}

func (e *Engine) Stop() error {
	e.done <- nil
	return <-e.done
}
