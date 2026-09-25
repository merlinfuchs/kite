-- name: CreateFlowAIPrompt :exec
INSERT INTO flow_ai_prompts (
    id,
    app_id,
    user_id,
    model,
    rounds,
    input_tokens,
    cached_input_tokens,
    output_tokens,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
);

-- name: GetFlowAIPrompt :one
SELECT * FROM flow_ai_prompts WHERE id = @id AND app_id = @app_id;

-- name: StartFlowAIPromptRound :execrows
UPDATE flow_ai_prompts SET
    rounds = rounds + 1,
    updated_at = @updated_at
WHERE id = @id AND app_id = @app_id AND rounds < @max_rounds;

-- name: AddFlowAIPromptUsage :exec
UPDATE flow_ai_prompts SET
    input_tokens = input_tokens + @input_tokens,
    cached_input_tokens = cached_input_tokens + @cached_input_tokens,
    output_tokens = output_tokens + @output_tokens,
    updated_at = @updated_at
WHERE id = @id AND app_id = @app_id;

-- name: DeleteFlowAIPrompt :exec
DELETE FROM flow_ai_prompts WHERE id = @id AND app_id = @app_id;

-- name: CountFlowAIPromptsByAppBetween :one
SELECT COUNT(*)::int FROM flow_ai_prompts WHERE app_id = @app_id AND created_at BETWEEN @start_at AND @end_at;
