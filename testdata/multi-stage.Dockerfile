FROM docker.io/alpine AS first_stage

CMD 'echo first stage'

FROM first_stage AS second_stage

CMD 'echo second stage'
