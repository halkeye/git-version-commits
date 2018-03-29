pipeline {
  agent {
    docker {
      image 'golang:1.9'
    }
  }

  options {
    timeout(time: 10, unit: 'MINUTES')
      ansiColor('xterm')
  }

  stages {
    stage('Build') {
      steps {
        dir("github.com/halkeye/git-version-commits") {
          checkout scm
        }
        sh "make"
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
