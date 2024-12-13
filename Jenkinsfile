pipeline {
    agent any
    
    environment {
        REPO = "SBT-DevOps-Weather-Reporter"
        GIT_HUB = "https://github.com/bonekeE/SBT-DevOps-Weather-Reporter.git/"
        DOCKER_IMAGE = "weather-reporter-service-server"
        DOCKER_TAG = "latest"
        DOCKER_HUB_LOGIN="bonekee"
    }

    stages {
        stage('Checkout') {
            steps {
                sh '''
                git clone $GIT_HUB
                '''
            }
        }

        stage('Build') {
            steps {
                script {
                    sh '''
                    cd $REPO && git checkout main
                    go mod tidy
                    go mod download
                    go build
                    '''
                }
            }
        }

        stage('Run Tests') {
            steps {
                script {
                    sh '''
                    cd $REPO
                    go test ./... -coverprofile=coverage.xml
                    '''
                }
            }
        }
        stage('SonarQube Analysis') {
            steps {
                withSonarQubeEnv('SonarQube') {
                    sh "${tool('SonarScanner')}/bin/sonar-scanner \
                    -Dsonar.projectKey=sbt_devops_project \
                    -Dsonar.projectName=weather_reporter_server \
                    -Dsonar.sources=$REPO/ \
                    -Dsonar.language=go \
                    -Dsonar.go.coverage.reportPaths=coverage.xml \
                    -Dsonar.host.url=http://81.94.150.203:9000 \
                    -Dsonar.login=squ_16acede43d8936420a730891917b830928af49be"
                }
            }
        }
        stage('Quality Gate') {
            steps {
                script {
                    timeout(time: 5, unit: 'MINUTES') {
                        waitForQualityGate abortPipeline: true
                    }
                }
            }
        }
        stage('Deploy') {
            steps {
                script {
                    sh '''
                    cd $REPO
            
                    docker build -t ${DOCKER_IMAGE}:${DOCKER_TAG} .
            
                    docker tag ${DOCKER_IMAGE}:${DOCKER_TAG} ${DOCKER_HUB_LOGIN}/${DOCKER_IMAGE}:${DOCKER_TAG}
                    docker push ${DOCKER_HUB_LOGIN}/${DOCKER_IMAGE}:${DOCKER_TAG}
                    
                    '''
                }
            }
        }
    }
        
    post {
        always {
            allure([
                reportBuildPolicy: 'ALWAYS',
                results: [[path: '$REPO/weather/allure-results']]
            ])
            sh 'rm -rf $REPO'
            sh 'docker rmi ${DOCKER_HUB_LOGIN}/${DOCKER_IMAGE}:${DOCKER_TAG}'
        }
        success {
            echo 'Pipeline finished successfully!'
        }

        failure {
            echo 'Pipeline failed!'
        }
    }
}

