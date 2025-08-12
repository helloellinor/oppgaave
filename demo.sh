#!/bin/bash

# ADHD Task Manager Demo Script
# This script demonstrates the core functionality of the task management system

echo "🧠 ADHD Task Management System Demo"
echo "=================================="

echo ""
echo "Starting server..."
go run main.go &
SERVER_PID=$!

# Wait for server to start
sleep 3

echo ""
echo "📊 Testing API endpoints..."

echo ""
echo "1. Getting all tasks:"
curl -s http://localhost:8080/api/tasks | jq -r '.[] | "- \(.title) (Priority: \(.priority), Cost: $\(.money_cost))"'

echo ""
echo "2. Creating a new task:"
NEW_TASK=$(curl -s -X POST -H "Content-Type: application/json" \
  -d '{"title":"Review Pull Requests","description":"Review team code submissions","estimated_duration_minutes":45,"priority":2,"energy_level":3,"difficulty":2}' \
  http://localhost:8080/api/tasks)

TASK_ID=$(echo $NEW_TASK | jq -r '.id')
echo "Created task #$TASK_ID: $(echo $NEW_TASK | jq -r '.title')"
echo "Money cost: $(echo $NEW_TASK | jq -r '.money_cost')"

echo ""
echo "3. Testing task dependencies:"
curl -s http://localhost:8080/api/tasks | jq -r '.[] | select(.prerequisites != null) | "- \(.title) depends on: \(.prerequisites[].title)"'

echo ""
echo "4. Budget Analysis:"
TOTAL_COST=$(curl -s http://localhost:8080/api/tasks | jq '[.[] | .money_cost] | add')
echo "Total pending task cost: $${TOTAL_COST}"
echo "Daily budget: $500"
echo "Remaining budget: $((500 - TOTAL_COST))"

if [ $((500 - TOTAL_COST)) -lt 0 ]; then
    echo "⚠️  OVER BUDGET!"
elif [ $((500 - TOTAL_COST)) -lt 100 ]; then
    echo "⚡ Running low on budget"
else
    echo "✨ Budget looks good"
fi

echo ""
echo "🌐 Dashboard available at: http://localhost:8080"
echo "📱 Try the web interface for full ADHD-friendly experience!"
echo ""
echo "Press Ctrl+C to stop the demo server..."

# Wait for user to stop
wait $SERVER_PID