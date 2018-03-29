pipeline {
  agent any

  options {
    timeout(time: 10, unit: 'MINUTES')
      ansiColor('xterm')
  }

  stages {
    stage('Build') {
      steps {
        dir("src/github.com/halkeye/git-version-commits") {
          checkout scm
          sh """
            export GOPATH=$WORKSPACE
            docker build -t halkeye/git-version-commits .
          """
        }
      }
    }
    stage('Deploy') {
      when {
        branch 'master'
      }
      environment {
        DOCKER = credentials('dockerhub-halkeye')
      }
      steps {
        sh 'docker login --username $DOCKER_USR --password=$DOCKER_PSW'
        sh 'docker push halkeye/git-version-commits'
      }
    }
  }
}
