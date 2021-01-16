pipeline {
    agent any
    stages {
        stage('build') {
            node {
                tools {
                    go 'go-1.15.6'
                }
                steps {
                    echo 'building...'
                    sh 'go build'
                }
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