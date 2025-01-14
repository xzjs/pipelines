# Copyright 2023 The Kubeflow Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
from kfp import compiler
from kfp import components
from kfp import dsl
from kfp.dsl import component


@component
def print_env_op():
    import os
    print('ENV1', os.environ.get('ENV1'))
    print('ENV2', os.environ.get('ENV2'))


print_env_2_op = components.load_component_from_text("""
name: Check env
implementation:
  container:
    image: alpine
    command:
    - sh
    - -c
    - |
      set -e -x
      if [ "$ENV2" == "val2" ]
      then
        echo "$ENV2" 
      else 
        echo "ENV2 does not equal val2"
        exit 1
      fi
      echo "$ENV3"
    env:
      ENV2: val0
""")


@dsl.pipeline(name='pipeline-with-env', pipeline_root='dummy_root')
def pipeline_with_env():
    print_env_op().set_env_variable(name='ENV1', value='val1')
    print_env_2_op().set_env_variable(
        name='ENV2', value='val2').set_env_variable(
            name='ENV3', value='val3')


if __name__ == '__main__':
    compiler.Compiler().compile(
        pipeline_func=pipeline_with_env,
        package_path='pipeline_with_env.yaml')