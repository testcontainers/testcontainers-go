import { check } from 'k6';
import http from 'k6/http';

export default function () {
  const res = http.get(`http://${__ENV.HTTPBIN}/status/200`);

  check(res, {
    'is status 200': (r) => r.status === 200,
  });
}