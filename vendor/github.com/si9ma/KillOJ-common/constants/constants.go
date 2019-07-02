package constants

import "time"

const ProjectName = "KillOJ"
const DefaultLang = "en"

// redis
const SubmitStatusKeyPrefix = "killoj_submit_status_"
const UserProblemSubmitIsCompletePrefix = "killoj_user_problem_submit_is_complete_"
const SubmitStatusTimeout = time.Hour // 1 hour

// sandbox
const JavaFile = "Main.java"
const GoFile = "main.go"
const CFile = "main.c"
const CppFile = "main.cpp"
const ExeFile = "Main"

// jwt
const JwtIdentityKey = "id"
