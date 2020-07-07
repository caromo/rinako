import sbt._

object Dependencies {
    val ackCordV = "0.16.1"

    val ackCord =       "net.katsstuff" %% "ackcord"                 % ackCordV
    val ackCordCore =   "net.katsstuff" %% "ackcord-core"            % ackCordV
    val ackCordComm =   "net.katsstuff" %% "ackcord-commands"        % ackCordV
    val ackCordLV =     "net.katsstuff" %% "ackcord-lavaplayer-core" % ackCordV

    val baseDeps = Seq(ackCord, ackCordCore, ackCordComm, ackCordLV)
}