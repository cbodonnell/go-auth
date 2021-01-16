node {
    // Ensure the desired Go version is installed
    def root = tool type: 'go', name: 'go1.15.6.linux-armv6l'

    // Export environment variables pointing to the directory where Go was installed
    withEnv(["GOROOT=${root}", "GOPATH=${root}/go", "PATH+GO=${root}/bin"]) {
        sh '$GOROOT/go/bin/go version'
        // stages {
        stage('build') {
            // steps {
                echo 'building...'
                sh '$GOROOT/go/bin/go build'
            // }
        }
        stage('test') {
            // steps {
                echo 'testing...'
            // }
        }
        stage('deploy') {
            // steps {
                echo 'deploying...'
                sh 'sudo cp go-auth /usr/local/bin/go-auth'
            // }
        }
        // }
        post {
            cleanup {
                deleteDir()
            }
        }
    }
}