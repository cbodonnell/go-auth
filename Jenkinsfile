pipeline {
    agent any
    environment {
        GOROOT = "${tool type: 'go', name: 'go1.15.6'}/go"
    }
    stages {
        stage('test') {
            steps {
                echo 'testing...'
                sh '$GOROOT/bin/go test github.com/cheebz/go-auth/hash'
            }
        }
        stage('build') {
            steps {
                echo 'building...'
                sh 'echo $GOROOT'
                sh '$GOROOT/bin/go build'
            }
        }
        stage('deploy') {
            steps {
                echo 'deploying...'
                sh 'sudo systemctl stop go-auth'
                sh 'sudo cp go-auth /etc/go-auth/go-auth'
                sh 'sudo cp -r templates/* /etc/go-auth/templates'
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