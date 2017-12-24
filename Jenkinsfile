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
    stage('Install') {
      steps {
        sh 'go get ./...'
      }
    }
    stage('Build') {
      steps {
        sh 'make all'
      }
    }
  }
}
