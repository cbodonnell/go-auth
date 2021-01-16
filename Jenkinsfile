pipeline {
    agent {
        // Run on an agent where we want to use Go
        node {
            // Ensure the desired Go version is installed
            def root = tool type: 'go', name: 'Go 1.15'

            // Export environment variables pointing to the directory where Go was installed
            withEnv(["GOROOT=${root}", "PATH+GO=${root}/bin"]) {
                sh 'go version'
            }
        }
    }
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