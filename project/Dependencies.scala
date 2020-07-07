import sbt._

object Dependencies {
  val ackCordV = "0.16.1"
  val akkaV = "2.6.4"

  val ackCord         =       "net.katsstuff" %% "ackcord"                 % ackCordV
  val ackCordCore     =   "net.katsstuff" %% "ackcord-core"            % ackCordV
  val ackCordComm     =   "net.katsstuff" %% "ackcord-commands"        % ackCordV
  val ackCordLV       =     "net.katsstuff" %% "ackcord-lavaplayer-core" % ackCordV
  val akkaActor       = "com.typesafe.akka"       %% "akka-actor"           % akkaV
  val akkaStream      = "com.typesafe.akka"       %% "akka-stream"          % akkaV
  val akkaRemote      = "com.typesafe.akka"       %% "akka-remote"          % akkaV
  val akkaTestkit     = "com.typesafe.akka"       %% "akka-testkit"         % akkaV
  val akkaSlf4j       = "com.typesafe.akka"       %% "akka-slf4j"           % akkaV


  val baseDeps = Seq(ackCord, ackCordCore, ackCordComm, ackCordLV, akkaActor, akkaStream, akkaRemote, akkaTestkit, akkaSlf4j)
}