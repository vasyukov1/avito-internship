import http from "k6/http";
import { check } from "k6";

export let options = {
    vus: 50,
    duration: "10s",
};

export function setup() {
    const teamPayload = {
        team_name: "team10",
        members: [
            { user_id: "u1", username: "Alice", is_active: true },
            { user_id: "u2", username: "Bob", is_active: true },
            { user_id: "u3", username: "Charlie", is_active: true }
        ]
    };

    http.post(
        "http://localhost:8080/team/add",
        JSON.stringify(teamPayload),
        { headers: { "Content-Type": "application/json" } }
    );

    const prId = `pr-${Date.now()}`;
    const payload = {
        pull_request_id: prId,
        pull_request_name: "Merge Test PR",
        author_id: "u1"
    };

    const res = http.post(
        "http://localhost:8080/pullRequest/create",
        JSON.stringify(payload),
        { headers: { "Content-Type": "application/json" } }
    );

    if (res.status !== 201) {
        console.error("PR creation failed:", res.body);
        return null;
    }

    const pr = res.json().pr;
    return { prID: pr.pull_request_id };
}

export default function (data) {
    if (!data || !data.prID) return;

    const mergeRes = http.post(
        "http://localhost:8080/pullRequest/merge",
        JSON.stringify({ pull_request_id: data.prID }),
        { headers: { "Content-Type": "application/json" } }
    );

    check(mergeRes, {
        "status is 200": (r) => r.status === 200,
        "PR is merged": (r) => {
            try {
                return r.json().pr.status === "MERGED";
            } catch {
                console.error("Failed to parse merge response:", r.body);
                return false;
            }
        }
    });
}
