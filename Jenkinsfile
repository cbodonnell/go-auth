pipeline {
    agent any
    // tools {
    //     go 'go1.15.6.linux-armv6l'
    // }
    environment {
        GOROOT = "${tool type: 'go', name: 'go1.15.6.linux-armv6l'}/go"
        // GOPATH
    }
    stages {
        stage('build') {
            steps {
                echo 'building...'
                sh 'echo $GOROOT'
                sh '$GOROOT/bin/go build'
            }
        }
        stage('test') {
            steps {
                echo 'testing...'
            }
        }
        stage('deploy') {
            steps {
                echo 'deploying...'
                sh 'sudo cp go-auth /home/pi/go/src/go-auth/go-auth'
                sh 'sudo systemctl restart go-auth'
            }
        }
    }
    post {
        cleanup {
            deleteDir()
        }
    }
}