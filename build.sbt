import Dependencies._

ThisBuild / scalaVersion := "2.13.1"
ThisBuild / version := "0.1.0"
ThisBuild / organization := "net.caromo"

resolvers += Resolver.JCenterRepository
lazy val root = (project in file("."))
  .settings(
    name := "rinako",
    libraryDependencies ++= Dependencies.baseDeps
  ).settings(mainClass in (Compile, run) := Some("net.caromo.Rinako"))