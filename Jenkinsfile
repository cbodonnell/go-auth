pipeline {
    agent any
    // tools {
    //     go 'go1.15.6.linux-armv6l'
    // }
    environment {
        GOROOT = tool type: 'go', name: 'go1.15.6.linux-armv6l'
    }
    stages {
        stage('build') {
            steps {
                echo 'building...'
                sh 'echo $GOROOT'
                sh 'go build'
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
                sh 'sudo cp go-auth /usr/local/bin/go-auth'
            }
        }
    }
    post {
        cleanup {
            deleteDir()
        }
    }
}