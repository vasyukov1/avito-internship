import http from "k6/http";
import { check, sleep } from "k6";

export let options = {
    vus: 100,
    duration: "30s",
    thresholds: {
        "http_req_duration": ["p(95)<300"],
        "http_req_failed": ["rate<0.001"]
    }
};

const BASE = "http://localhost:8080";

const TEST_USERS = Array.from({length: 10}, (_, i) => `load_user_${i+1}`);

export function setup() {
    const members = TEST_USERS.map((id, idx) => {
        return {
            user_id: id,
            username: `LoadUser${idx+1}`,
            is_active: true
        };
    });

    const teamBody = JSON.stringify({
        team_name: "load_backend",
        members: members
    });

    const res = http.post(`${BASE}/team/add`, teamBody, { headers: { "Content-Type": "application/json" } });

    if (res.status !== 201 && res.status !== 400) {
        console.error("Failed to create team in setup:", res.status, res.body);
    } else {
        console.log("Team created/exists, setup ok:", res.status);
    }

    return { authors: TEST_USERS };
}

export default function(data) {
    const authors = data.authors;
    const author = authors[Math.floor(Math.random() * authors.length)];

    const prID = `pr-load-${__VU}-${__ITER}-${Date.now()}`;

    const payload = JSON.stringify({
        pull_request_id: prID,
        pull_request_name: `Load Test PR ${__VU}-${__ITER}`,
        author_id: author
    });

    const res = http.post(
        `${BASE}/pullRequest/create`,
        payload,
        { headers: { "Content-Type": "application/json" } }
    );

    const ok = check(res, {
        "created status is 201": (r) => r.status === 201,
        "response has pr": (r) => {
            try {
                const j = r.json();
                return j && j.pr && j.pr.pull_request_id === prID;
            } catch (e) {
                return false;
            }
        }
    });

    if (!ok) {
        console.error("Create PR failed:", res.status, res.body);
    }

    sleep(0.05);
}
