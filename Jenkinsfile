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
          sh """
            export GOPATH=$WORKSPACE
            go get ./...
            mkdir -p bin
            for binary in confluence-poster git-release-info release-info-confluence; do
              for OS in darwin linux windows; do
                env CGO_ENABLED=0 GOOS=\$OS GOARCH=amd64 go build -a -installsuffix cgo -o bin/\${binary}-\${OS}-amd64 \${binary}/main.go; \
              done;
            done
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
