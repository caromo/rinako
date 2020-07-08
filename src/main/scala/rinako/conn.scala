package net.caromo

import ackcord._
import ackcord.data._

import com.typesafe.config._
import akka.actor.{ActorSystem, Props}

import scala.util.{Failure, Success}
import scala.concurrent.ExecutionContext.Implicits.global
import scala.concurrent._
import scala.concurrent.duration._

object Rinako extends App {

  // Read config
  val config = ConfigFactory.load()
  val rinakoConf = config.getConfig("rinako")

  setupClient(rinakoConf).createClient().foreach { cli =>
    cli.onEventSideEffectsIgnore {
      case APIMessage.Ready(_) => println("Now ready")
    }

    val textListener = new TextListener(cli)
    val plainCommands = new PlainCommands(cli, cli.requests)

    cli.registerListener(textListener.init)
    cli.commands.bulkRunNamed(
      plainCommands.greet
    )

    cli.login
  }
  
  def setupClient(config: Config): ClientSettings = ClientSettings(config.getString("token"))
}