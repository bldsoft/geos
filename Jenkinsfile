def APP_VERSION = ''
def APP_REVISION = ''
def IMAGE_NAME = ''



def STAGE_NAME = 'Build Geos'
def JOB_NAME = 'Setrix/Geos/Build'
def SLACK_CHANNEL = 'setrix_build'
def COLOR_MAP = [
    'SUCCESS': 'good',
    'FAILURE': 'danger',
    'ABORTED': 'warning',
]


def getBuildUser() {
    if (currentBuild.rawBuild.getCause(Cause.UserIdCause)) {
        return currentBuild.rawBuild.getCause(Cause.UserIdCause).getUserId()
    } else {
        return false
    }
}


pipeline {
    agent any

    stages {

        stage('Checkout sources from GitHub') {
            steps {
                script {
		            APP_VERSION = readFile(file: "VERSION")
                    echo "Building app Geos version: ${APP_VERSION}"
                }
                script {
                    APP_REVISION = "${GIT_COMMIT}".substring(0,8)
                    echo "Building app revision: ${APP_REVISION}"
                }
                script {
                    IMAGE_NAME = "setplexapps/geos:${APP_VERSION}-${APP_REVISION}"
                    echo "Image name: ${IMAGE_NAME} "
                }
            }
        }

        stage('Build and push application docker image') {
            steps {
                script {
                    docker.withRegistry('','dockerhub-push') {
                        def dockerImage = docker.build("${IMAGE_NAME}",
                        "--build-arg COMMITHASH=${APP_REVISION} --build-arg VERSION=${APP_VERSION} --build-arg DATE=${BUILD_TIMESTAMP} -f Dockerfile ./")
                        dockerImage.push()
                    }
                }
            }
        }

        stage('Create buildinfo.json artifact') {
            steps {
                script {
                    def buildInfo = [dockerImageName:"setplexapps/geos:${APP_VERSION}-${APP_REVISION}"]
                    writeJSON file: 'buildInfo.json', json: buildInfo
                    archiveArtifacts 'buildInfo.json'
                }
	          }
	      }
    }
	
	
    post {
        always {
            script {
                def slack_message = "${currentBuild.currentResult}:\nProject: `Geos`\nStage: `${STAGE_NAME}`\nJob: `${JOB_NAME}`\nApp Version: `${APP_VERSION}-${APP_REVISION}`"

                if (getBuildUser()){
                    slack_message +=  "\nBy user: `${getBuildUser()}`"
                }

                slackSend(
                    channel: SLACK_CHANNEL,
                    color: COLOR_MAP[currentBuild.currentResult],
                    message: slack_message
                )
            }
        }
    }	
}
