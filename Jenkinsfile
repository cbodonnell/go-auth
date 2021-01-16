pipeline {
    agent any
    stages {
        stage('build') {
            steps {
                echo 'building...'
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
                sh 'sudo cp go-auth /home/craig/go/src/go-auth/go-auth'
            }
        }
    }
    post {
        cleanup {
            deleteDir()
        }
    }
}