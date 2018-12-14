package test

import (
	"github.com/teivah/gosiris/gosiris"
	"testing"
	"time"
)

func TestTry(t *testing.T) {
	//Init a local actor system
	gosiris.InitActorSystem(gosiris.SystemOptions{
		ActorSystemName: "ActorSystem",
	})

	//Create an actor
	parentActor := gosiris.Actor{}
	//Close an actor
	defer parentActor.Close()

	//Create an actor
	childActor := gosiris.Actor{}
	//Close an actor
	defer childActor.Close()
	//Register a reaction to event types ("message" in this case)
	childActor.React("message", func(context gosiris.Context) {
		context.Self.LogInfo(context, "Received %v\n", context.Data)
	})

	//Register an actor to the system
	gosiris.ActorSystem().RegisterActor("parentActor", &parentActor, nil)
	//Register an actor by spawning it
	gosiris.ActorSystem().SpawnActor(&parentActor, "childActor", &childActor, nil)

	//Retrieve actor references
	parentActorRef, _ := gosiris.ActorSystem().ActorOf("parentActor")
	childActorRef, _ := gosiris.ActorSystem().ActorOf("childActor")

	//Send a message from one actor to another (from parentActor to childActor)
	childActorRef.Tell(gosiris.EmptyContext, "message", "Hi! How are you?", parentActorRef);

	time.Sleep(time.Second * 600);
}