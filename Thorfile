require 'thor'
require 'thor/actions'
require 'thor/scmversion'
require 'octokit'

## GitHub Helpers
module GitHub
  class << self
    def client
      raise 'Missing required environment variable GITHUB_TOKEN' unless ENV.include?('GITHUB_TOKEN')

      @client ||= Octokit::Client.new(:access_token => ENV['GITHUB_TOKEN'])
    end

    def version
      IO.read(File.join(__dir__, 'VERSION'))
    end

    def repo
      ENV['TRAVIS_REPO_SLUG']
    end

    def commit
      ENV['TRAVIS_COMMIT']
    end
  end

  ## Thor Commands
  class Commands < Thor
    namespace 'gh'

    desc 'release BUILD_DIR=build', 'Upload all artifacts in BUILD_DIR to the current GitHub release'
    def release(build_dir = 'build')
      say_status :tag, "Creating draft release #{GitHub.version} on #{GitHub.repo}"
      handle = GitHub.client.create_release(GitHub.repo, GitHub.version,
                                            :target_commitish => GitHub.commit,
                                            :name => "aws-verify #{GitHub.version}",
                                            :draft => true)

      Dir[File.join(build_dir, '*')].each do |artifact|
        say_status :upload, artifact
        GitHub.client.upload_asset(handle.url, artifact,
                                   :name => File.basename(artifact))
      end
    end
  end
end

module Gox
  GOARCH = %w(amd64 386 arm).freeze
  GOOS = %w(darwin linux windows).freeze

  ## Thor Commands
  class Commands < Thor
    include Thor::Actions

    namespace 'gox'

    desc 'build BUILD_DIR=build', 'Perform a gox build, placing artifacts into BUILD_DIR'
    def build(build_dir = 'build')
      empty_directory build_dir
      run [
        'gox',
        "-output=#{build_dir}/{{.Dir}}-#{GitHub.version}-{{.OS}}-{{.Arch}}",
        %(-arch="#{GOARCH.join(' ')}"),
        %(-os="#{GOOS.join(' ')}")
      ].join(' ')
    end
  end
end
