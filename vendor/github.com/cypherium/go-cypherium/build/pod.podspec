Pod::Spec.new do |spec|
  spec.name         = 'Cypher'
  spec.version      = '{{.Version}}'
  spec.license      = { :type => 'GNU Lesser General Public License, Version 3.0' }
  spec.homepage     = 'https://github.com/cypherium/go-cypherium'
  spec.authors      = { {{range .Contributors}}
		'{{.Name}}' => '{{.Email}}',{{end}}
	}
  spec.summary      = 'iOS Cypherium Client'
  spec.source       = { :git => 'https://github.com/cypherium/go-cypherium.git', :commit => '{{.Commit}}' }

	spec.platform = :ios
  spec.ios.deployment_target  = '9.0'
	spec.ios.vendored_frameworks = 'Frameworks/Cypher.framework'

	spec.prepare_command = <<-CMD
    curl https://gcphstore.blob.core.windows.net/builds/{{.Archive}}.tar.gz | tar -xvz
    mkdir Frameworks
    mv {{.Archive}}/Cypher.framework Frameworks
    rm -rf {{.Archive}}
  CMD
end
