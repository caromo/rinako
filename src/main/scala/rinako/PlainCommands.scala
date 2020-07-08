package net.caromo

import ackcord._
import ackcord.data._
import ackcord.commands._
import ackcord.syntax._
import akka.NotUsed

class PlainCommands(cli: DiscordClient, req: Requests) extends CommandController(req) {
  val discriminator = "!"
  
  import MessageParser.Auto._

  val greet: NamedComplexCommand[String, NotUsed] = Command
    .named(discriminator, Seq("rinako"), mustMention = false)
    .parsing[String]
    .withRequest(m => {println("received"+m); m.textChannel.sendMessage(s"Rina-chan Board 'mu ${m.user.username}'")})
}