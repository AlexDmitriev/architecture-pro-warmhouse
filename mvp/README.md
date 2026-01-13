- heath 
  - event listener (2 events - change status) + log (вместо записи в бд) 
  - switch rest apt (запись в БД)
  - отправка сообщения в сервис телеметрии 

- light
  - event listener (2 events) + log (вместо записи в бд)
  - set (запись в БД)
  - отправка сообщения в сервис телеметрии

- gates
  - event listener (2 events) + log (вместо записи в бд)
  - set (запись в БД)
  - отправка сообщения в сервис телеметрии

- telemetry 
 - event listener (3 type of events) + log (запись в бд)
 - rest api для просмотра телеметрии по устройству


База данных 

locations
 - id 
 - name
 
sensors
 - id
 - location_id
 - name
 - serial_number
 - status
 - created_at
 - last_activity_at

sensor_types
 - id
 - name
 - code 

sensor_heating
 - id 
 - serial_number
 - temperature
 - created_at
 - updated_at

sensor_lighting
 - id
 - serial_number
 - status
 - created_at
 - updated_at

sensor_gates
 - id
 - serial_number
 - status
 - created_at
 - updated_at

telemetry
 - id 
 - serial_number
 - value
 - created_at