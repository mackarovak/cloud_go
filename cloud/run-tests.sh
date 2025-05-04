#!/bin/sh

# Ждем готовности сервисов
echo "Waiting for services to be ready..."
timeout 30 sh -c 'until curl -f http://backend1:80/health && curl -f http://backend2:80/health && curl -f http://load-balancer:8080/health; do 
  sleep 1
done'

if [ $? -ne 0 ]; then
  echo "Services failed to start in time"
  exit 1
fi

# Запускаем тесты
echo "Running tests..."
./app.test -test.v \
  -test.parallel ${TEST_PARALLEL:-4} \
  -test.timeout 10m \
  -test.coverprofile=/app/test_results/coverage.out \
  -test.bench=. \
  -test.benchmem \
  -test.benchtime=5s

# Генерируем отчет о покрытии
go tool cover -html=/app/test_results/coverage.out -o /app/test_results/coverage.html

# Сохраняем результат
jq -n --arg exitCode "$?" '{exit_code: $exitCode|tonumber}' > /app/test_results/result.json
exit $?