package net.caromo

import ackcord._
import ackcord.data._
import scala.concurrent._
import scala.concurrent.duration._

object Rinako extends App {

  val token = "<token>"

  val clientSettings = ClientSettings(token)
  import clientSettings.executionContext

  val client = Await.result(clientSettings.createClient(), Duration.Inf)

  client.onEventSideEffectsIgnore {
    case APIMessage.Ready(_) => println("Now ready")
  }

  client.login
}