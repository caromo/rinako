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

  val fCli = setupClient(rinakoConf)

  fCli onComplete {
    case Success(cli) =>
      cli.onEventSideEffectsIgnore {
        case APIMessage.Ready(_) => println("Now ready")
      }
      cli.login
    case Failure(e) => 
      e.printStackTrace()
  }
  
  def setupClient(config: Config): Future[DiscordClient] = {
    val token = config.getString("token")
    val clientSettings = ClientSettings(token)
    clientSettings.createClient
  }
}