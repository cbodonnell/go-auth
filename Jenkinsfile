pipeline {
    agent any
    environment {
        GOROOT = "${tool type: 'go', name: 'go1.15.6.linux-armv6l'}/go"
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
                sh 'sudo systemctl stop go-auth'
                sh 'sudo cp go-auth /etc/go-auth/go-auth'
                sh 'sudo cp -r templates /etc/go-auth/templates'
                sh 'sudo systemctl start go-auth'
            }
        }
    }
    post {
        cleanup {
            deleteDir()
        }
    }
}