# Base object
___
***Note***:
    - id (uuid)
    - title
    - description
    - status (pending, done, canceled)
    - priority (low, medium, high)
    - created_at
    - updated_at

# Endpoints
___
1. POST `api/v1/note` - creating note
2. PATCH `api/v1/note` - updating note
3. GET `api/v1/note/:id` - getting note
4. GET `api/v1/notes?page=&limit=` - getting note list
5. DELETE `api/v1/note/:id` - deleting note