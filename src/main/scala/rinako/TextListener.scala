package net.caromo

import ackcord._
import ackcord.data._
import akka.NotUsed

class TextListener(cli: DiscordClient) extends EventsController(cli.requests) {

  val MessageEvent: EventListenerBuilder[TextChannelEventListenerMessage, APIMessage.MessageCreate] =
    TextChannelEvent.on[APIMessage.MessageCreate]
  val defaultDisc = ">"

  // def listen(discriminator: String, channelID: TextChannelId): EventListener[APIMessage.MessageCreate, NotUsed] = 
  //   MessageEvent.withSideEffects { msg => 
  //     if (msg.channel.id == channelID) {
  //       println(s"$discriminator: ${msg.event.message.content}")
  //     }
  //   }

  // def stop(discriminator: String, channelID: TextChannelId, listener: EventRegistration[NotUsed], stopper: EventRegistration[NotUsed]): EventListener[APIMessage.MessageCreate, NotUsed] = {
  //   MessageEvent.withSideEffects { msg =>
  //     if (msg.channel.id == channelID) {
  //       listener.stop
  //       stopper.stop
  //     }
  //   }
  // }

  val init: EventListener[APIMessage.MessageCreate, NotUsed] =
    MessageEvent.withSideEffects { me =>
      val msg = me.event.message

      if (msg.content.startsWith(defaultDisc)) {
        val cid = msg.channelId
        // val listener = cli.registerListener(listen(defaultDisc, cid))
        val strippedMsg = msg.content.replaceFirst(defaultDisc, "")
        val command = strippedMsg.takeWhile(_ != ' ')
        val args = strippedMsg.replaceFirst(command, "")

        // lazy val stopper: EventRegistration[NotUsed] = 
        //   cli.registerListener(stop(defaultDisc, cid, listener, stopper))

        // stopper
      }

    }
}