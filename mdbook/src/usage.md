# Usage

## Start the Server:

First, start the server:

```sh
docker compose up

# Or without docker
draftsmith_api serve
```

## List of Endpoints
The following endpoints are provided, with `POST`, `PUT`, `GET` and `DELETE`, implementations as described below:

- /notes
    - /notes/no-content
    - /notes/search
    - /notes/tree
    - /notes/{id}
    - /notes/{id}/tags
- /tags
    - /tags/tree
    - /tags/with-notes
    - /tags/{id}
- /task_clocks
    - /task_clocks/{id}
- /task_schedules
    - /task_schedules/{id}
- /tasks
    - /tasks/details
    - /tasks/tree
    - /tasks/{id}

## API Documentation

### Checklist

- Notes
    - [x] Create
    - [x] Update
    - [x] Delete
    - [x] Get
    - [x] Search
    - hierarchy
        - [x] Create
        - [x] Get (Tree)
        - [x] Delete
        - [x] Update
- Tags
    - [x] Create
    - [x] Update
    - [x] Delete
    - [x] Assign
    - [x] Get
    - [-] Search
        - As in Search notes assigned to a tag
            - Left to the client to filter out typical search results
    - [-] Filter
        - Left to the client to user a `fzf` tool
    - hierarchy
        - [x] Create
        - [x] Get (Tree)
        - [x] Delete
        - [x] Update
- Tasks
    - Assign
       - [x] Create
       - [x] Delete
       - [x] Update
       - [x] Get
         - Getting Hierarchical tasks will involve merging in the details from the flat `/details` method.
    - Schedule
        - [x] Create
        - [x] Delete
        - [x] Update
        - [-] Get
            - Included in the `tasks/details` endpoint.
    - Clock in
        - [x] Create
        - [x] Delete
        - [x] Update
        - [x] Get
            - Included in the `tasks/details` endpoint.
- Categories
    - Being removed




### Notes
#### Discussion
##### Flat
Notes are stored in a table `notes` like so:

```
-[ RECORD 1 ]--------------------------------------------
id          | 1
title       | First note
content     | This is the first note in the system.
created_at  | 2024-10-20 05:04:42.709064
modified_at | 2024-10-20 05:04:42.709064
fts         | 'first':1,6 'note':2,7 'system':10
-[ RECORD 2 ]--------------------------------------------
id          | 2
title       | Foo
content     | This is the updated content of the note.
created_at  | 2024-10-20 05:04:42.709064
modified_at | 2024-10-20 05:15:03.334779
fts         | 'content':6 'foo':1 'note':9 'updat':5
-[ RECORD 3 ]--------------------------------------------
id          | 3
title       | Foo
content     | Bar
created_at  | 2024-10-20 05:20:20.938922
modified_at | 2024-10-20 05:20:20.938922
fts         | 'bar':2 'foo':1
-[ RECORD 4 ]--------------------------------------------
id          | 4
title       | New Note Title
content     | This is the content of the new note.
created_at  | 2024-10-20 05:20:43.792369
modified_at | 2024-10-20 05:20:43.792369
fts         | 'content':7 'new':1,10 'note':2,11 'titl':3

draftsmith=#
```
##### hierarchy
There is a table `note_hierarchy` that stores the hierarchy of notes. The `hierarchy_type` column can be `subpage` or `section`:

```
select * from note_hierarchy;
 id | parent_note_id | child_note_id | hierarchy_type
----+----------------+---------------+----------------
  1 |              1 |             2 | subpage
  2 |              1 |             2 | subpage
(2 rows)
```
#### Flat
##### Create

```sh
curl -X POST http://localhost:37238/notes \
      -H "Content-Type: application/json" \
      -d '{"title": "New Note Title", "content": "This is the content of the new note."}'

```

```json
{"id":4,"message":"Note created successfully"}
```
##### Update
To update the title of note 1:
```sh
curl -X PUT -H "Content-Type: application/json" -d '{"title":"New Title"}' http://localhost:37238/notes/1
```
To update the content:

```sh
curl -X PUT -H "Content-Type: application/json" -d '{"content":"New content"}' http://localhost:37238/notes/1
```

##### Delete

```sh
curl -X DELETE http://localhost:37238/notes/6
```

```json
{"message":"Note deleted successfully"}
```

##### Get
###### All

```
curl http://localhost:37238/notes | jq
```

```json
[
  {
    "id": 1,
    "title": "First note",
    "content": "This is the first note in the system.",
    "created_at": "2024-10-20T05:04:42.709064Z",
    "modified_at": "2024-10-20T05:04:42.709064Z"
  },
  {
    "id": 2,
    "title": "Foo",
    "content": "This is the updated content of the note.",
    "created_at": "2024-10-20T05:04:42.709064Z",
    "modified_at": "2024-10-20T05:15:03.334779Z"
  },
  {
    "id": 3,
    "title": "Foo",
    "content": "Bar",
    "created_at": "2024-10-20T05:20:20.938922Z",
    "modified_at": "2024-10-20T05:20:20.938922Z"
  },
  {
    "id": 4,
    "title": "New Note Title",
    "content": "This is the content of the new note.",
    "created_at": "2024-10-20T05:20:43.792369Z",
    "modified_at": "2024-10-20T05:20:43.792369Z"
  }
]

```

It's also possible to get notes without their content (useful for palettes etc.):

```sh
curl http://localhost:37238/notes/no-content
```

```json
[
  {
    "id": 1,
    "title": "First note",
    "created_at": "2024-10-20T05:04:42.709064Z",
    "modified_at": "2024-10-20T05:04:42.709064Z"
  },
  {
    "id": 2,
    "title": "Foo",
    "created_at": "2024-10-20T05:04:42.709064Z",
    "modified_at": "2024-10-20T05:15:03.334779Z"
  },
  {
    "id": 3,
    "title": "Foo",
    "created_at": "2024-10-20T05:20:20.938922Z",
    "modified_at": "2024-10-20T05:20:20.938922Z"
  },
  {
    "id": 4,
    "title": "New Note Title",
    "created_at": "2024-10-20T05:20:43.792369Z",
    "modified_at": "2024-10-20T05:20:43.792369Z"
  }
]
```
##### Search
If using curl, make sure to handle spaces in the query string:

```sh
curl "http://localhost:37238/notes/search?q=updated%20note"
```

Or use Python:

```python
import requests
from urllib.parse import quote
import json

query = quote("updated content")
response = requests.get(f"http://localhost:37238/notes/search?q={query}")
# output = json.loads(response.text)
output = response.json()
print(json.dumps(output, indent=2))
```

```json
[
  {
    "id": 2,
    "title": "Foo"
  }
]
```
#### hierarchy
##### Examples
Consider some notes:

```sh
curl localhost:37238/notes/tree | jq
```

```json
[
  {
    "id": 1,
    "title": "New Title",
    "type": "",
    "children": [
      {
        "id": 2,
        "title": "Foo",
        "type": "subpage",
        "children": [
          {
            "id": 3,
            "title": "Foo",
            "type": "subpage",
            "children": [
              {
                "id": 4,
                "title": "New Note Title",
                "type": "subpage"
              }
            ]
          }
        ]
      }
    ]
  }
]
```
These can be flattened:

```sh
curl -X DELETE http://localhost:37238/notes/hierarchy/4
```
```json
[
  {
    "id": 4,
    "title": "New Note Title",
    "type": ""
  },
  {
    "id": 1,
    "title": "New Title",
    "type": "",
    "children": [
      {
        "id": 2,
        "title": "Foo",
        "type": "subpage",
        "children": [
          {
            "id": 3,
            "title": "Foo",
            "type": "subpage"
          }
        ]
      }
    ]
  }
]
```

```sh
curl -X DELETE http://localhost:37238/notes/hierarchy/3
curl -X DELETE http://localhost:37238/notes/hierarchy/2
```

```json
[
  {
    "id": 3,
    "title": "Foo",
    "type": ""
  },
  {
    "id": 4,
    "title": "New Note Title",
    "type": ""
  },
  {
    "id": 1,
    "title": "New Title",
    "type": ""
  },
  {
    "id": 2,
    "title": "Foo",
    "type": ""
  }
]

```

Then they can be re-attached:

```python
import requests

# Define the URL endpoint
url = "http://localhost:37238/notes/hierarchy"

# Define the headers
headers = {
    "Content-Type": "application/json"
}

for i in range(1, 4):


    # Define the payload
    payload = {
        "parent_note_id": i,
        "child_note_id": i+1,
        "hierarchy_type": "subpage"
    }

    # Make the POST request
    response = requests.post(url, headers=headers, json=payload)

    # Check the response
    print(response.status_code)
    print(response.json())

```

```json
[
  {
    "id": 1,
    "title": "New Title",
    "type": "",
    "children": [
      {
        "id": 2,
        "title": "Foo",
        "type": "subpage",
        "children": [
          {
            "id": 3,
            "title": "Foo",
            "type": "subpage",
            "children": [
              {
                "id": 4,
                "title": "New Note Title",
                "type": "subpage"
              }
            ]
          }
        ]
      }
    ]
  }
]
```

Notes can also be detached separately:

```sh
curl -X DELETE http://localhost:37238/notes/hierarchy/2
```

```json
[
  {
    "id": 2,
    "title": "Foo",
    "type": "",
    "children": [
      {
        "id": 3,
        "title": "Foo",
        "type": "subpage",
        "children": [
          {
            "id": 4,
            "title": "New Note Title",
            "type": "subpage"
          }
        ]
      }
    ]
  },
  {
    "id": 1,
    "title": "New Title",
    "type": ""
  }
]
```
To move 4 under 2:

```sh
curl -X PUT http://localhost:37238/notes/hierarchy/4 \
      -H "Content-Type: application/json" \
      -d '{"parent_note_id": 2, "hierarchy_type": "subpage"}'


```

```
[
  {
    "id": 1,
    "title": "New Title",
    "type": ""
  },
  {
    "id": 2,
    "title": "Foo",
    "type": "",
    "children": [
      {
        "id": 3,
        "title": "Foo",
        "type": "subpage",
        "children": [
          {
            "id": 4,
            "title": "New Note Title",
            "type": "subpage"
          }
        ]
      }
    ]
  }
]
```
We can then update that so 4 is under 2:

```sh
curl -X PUT http://localhost:37238/notes/hierarchy/4 \
        -H "Content-Type: application/json" \
        -d '{"parent_note_id": 2, "hierarchy_type": "subpage"}'
```
```sh
curl http://localhost:37238/notes/tree | jq
```

```json
[
  {
    "id": 2,
    "title": "Foo",
    "type": "",
    "children": [
      {
        "id": 3,
        "title": "Foo",
        "type": "subpage"
      },
      {
        "id": 4,
        "title": "New Note Title",
        "type": "subpage"
      }
    ]
  },
  {
    "id": 1,
    "title": "New Title",
    "type": ""
  }
]
```

**NOTE**, be careful here, because the server will check for cycles but nothing is perfect. For example, if we reset them:

```sh
fish
for i in (seq 4)
    curl -X DELETE http://localhost:37238/notes/hierarchy/$i
end

# Run python script above
```

```json
[
  {
    "id": 1,
    "title": "New Title",
    "type": "",
    "children": [
      {
        "id": 2,
        "title": "Foo",
        "type": "subpage",
        "children": [
          {
            "id": 3,
            "title": "Foo",
            "type": "subpage",
            "children": [
              {
                "id": 4,
                "title": "New Note Title",
                "type": "subpage"
              }
            ]
          }
        ]
      }
    ]
  }
]
```

Now If we move 2 under 4:

```sh
curl -X PUT http://localhost:37238/notes/hierarchy/2 \
        -H "Content-Type: application/json" \
        -d '{"parent_note_id": 4, "hierarchy_type": "subpage"}'

curl http://localhost:37238/notes/tree | jq
```

```json

[
  {
    "id": 1,
    "title": "New Title",
    "type": ""
  }
]
```


Then try to move 2 under 4, we would have this:

```
draftsmith=# select * from note_hierarchy;
 id | parent_note_id | child_note_id | hierarchy_type
----+----------------+---------------+----------------
 21 |              2 |             3 | subpage
 22 |              3 |             4 | subpage
 20 |              4 |             2 | subpage
```

Map that out:

- 2 -> 3 -> 4 -> 2
    - 2 -> 3 -> 4 -> 2
        - 2 -> 3 -> 4 -> 2


The server will check this (as of commit `53c13f82`), the client should also take care not to rely on this.


##### Create

```sh
curl -X POST http://localhost:37238/notes/hierarchy \
      -H "Content-Type: application/json" \
      -d '{"parent_note_id": 1, "child_note_id": 2, "hierarchy_type": "subpage"}'

```
```json
{"id":2,"message":"Note hierarchy entry added successfully"}
```
##### Update
Clients should use the update endpoint to change the type of hierarchy (e.g. from `subpage` to `section`) or to change the parent note. This is better than add and delete because it is atomic (and only one request is needed).

```sh
curl -X PUT http://localhost:37238/notes/hierarchy/4 \
      -H "Content-Type: application/json" \
      -d '{"parent_note_id": 2, "hierarchy_type": "subpage"}'
```
```
j"message":"Note hierarchy entry updated successfully"}
```


##### Delete
A note can only have one parent, so deleting a hierarchy tags the child tag as an argument.

For example to remove to from whatever parent it has:

```sh
curl -X DELETE http://localhost:37238/notes/hierarchy/2
```
```
{"message":"Note hierarchy entry deleted successfully"}
```
##### Get (Tree)
```sh
curl http://localhost:37238/notes/tree
```

```json
[
  {
    "id": 3,
    "title": "Foo",
    "type": ""
  },
  {
    "id": 4,
    "title": "New Note Title",
    "type": ""
  },
  {
    "id": 1,
    "title": "First note",
    "type": "",
    "children": [
      {
        "id": 2,
        "title": "Foo",
        "type": "subpage"
      },
      {
        "id": 2,
        "title": "Foo",
        "type": "subpage"
      }
    ]
  }
]
```
### Assets
The assets are stored in a `./uploads` directory. Each file has an entry in the database:

```sql
SELECT * FROM ASSETS;
```
```
draftsmith=# select * from assets;
 id | note_id | asset_type |       location        |          description          |         created_at
----+---------+------------+-----------------------+-------------------------------+----------------------------
  2 |       1 |            | uploads/asldkfj.org   | This is a sample image upload | 2024-10-26 07:14:01.973793
  3 |       1 |            | uploads/asldkfj.org   | This is a sample image upload | 2024-10-26 07:14:01.973793
  4 |       1 |            | uploads/asldkfj.org   | This is a sample image upload | 2024-10-26 07:14:01.973793
  5 |       1 |            | uploads/asldkfj.org   | This is a sample image upload | 2024-10-26 07:14:01.973793
  6 |       1 |            | uploads/asldkfj.org   | This is a sample image upload | 2024-10-26 07:14:01.973793
  7 |       1 |            | uploads/asldkfj.org   | This is a sample image upload | 2024-10-26 07:14:01.973793
  8 |       1 |            | uploads/asldkfj_1.org | This is a sample image upload | 2024-10-26 07:14:01.973793
  9 |       1 |            | uploads/asldkfj_2.org | This is a sample image upload | 2024-10-26 07:14:01.973793
 10 |       1 |            | uploads/asldkfj_3.org |                               | 2024-10-26 07:14:01.973793
 11 |       1 |            | uploads/asldkfj.org   |                               | 2024-10-26 07:14:01.973793
 12 |       1 |            | uploads/asldkfj_4.org |                               | 2024-10-26 07:14:01.973793
 15 |         |            | uploads/asldkfj_5.org |                               | 2024-10-26 07:14:01.973793
 16 |         |            | uploads/asldkfj_6.org |                               | 2024-10-26 07:14:01.973793
(13 rows)
```

The database tracks the content primarily for the `note_id` and `description` fields:

  - `note_id`
      - Not Implemented
          - A page that will represent this asset, much like mediawiki it can serve as a place to serve notes, citations or code used to generate it (e.g. python / R plots or TikZ / Asymptote diagrams).
  - `description`
      - Not Implemented
          - This will be used for search, either by running `pdftotext` on the file or using LLava to automatically describe the images.

The `location` field is the path to the file in the `./uploads` directory.

#### Create (POST)

```bash
curl -X POST \
   http://localhost:37238/upload \
   -H "Content-Type: multipart/form-data" \
   -F "file=@/tmp/asldkfj.org" \
   -F "note_id=1" \
   -F "description=This is a sample image upload"


```

One can upload without a description:

```bash
curl -X POST \
   http://localhost:37238/upload \
   -H "Content-Type: multipart/form-data" \
   -F "file=@/tmp/asldkfj.org" \
   -F "note_id=1" \
   -F "description=This is a sample image upload"
```

```
{"filename":"asldkfj.org","id":1,"message":"File uploaded successfully"}
```

Without a description:


```bash
   http://localhost:37238/upload \
   -H "Content-Type: multipart/form-data" \
   -F "file=@/path/to/your/file.txt" \
   -F "note_id=3"
```

```
{"filename":"asldkfj_2.org","id":2,"message":"File uploaded successfully"}
```

This will simply upload the file into a `./uploads` directory from wherever the binary is run.

It won't clober files either, they will be renamed.

#### Delete (DELETE)
##### DB Entries with No Files
This is not implemented as I can foresee pain when people move around docker containers. Here's a python script that calls the `DELETE` method. The `DELETE` method succeeds even if the file is not found, the logic is:


1. GET request for database entries.
2. Check if each asset's `file_name` exists in the `./uploads` directory.
3. If the file does not exist, send a DELETE request to remove the asset from the db.

```python
import os
import requests

# Base URL for the API
base_url = 'http://localhost:37238/assets'

# Directory to check for file existence
uploads_directory = './uploads'

# Get the list of assets
response = requests.get(base_url)

if response.status_code == 200:
    assets = response.json()
    for asset in assets:
        file_name = asset['file_name']
        file_path = os.path.join(uploads_directory, file_name)

        # Check if file exists in the uploads directory
        if not os.path.isfile(file_path):
            # File does not exist, perform DELETE
            delete_url = f"{base_url}/{asset['id']}"
            delete_response = requests.delete(delete_url)

            if delete_response.status_code == 200:
                print(f"Deleted asset with ID: {asset['id']}")
            else:
                print(f"Failed to delete asset with ID: {asset['id']}. Status code: {delete_response.status_code}")
else:
    print(f"Failed to retrieve assets. Status code: {response.status_code}")
```

##### Files with No DB Entry
The cleanupOrphanedFiles function will run every 24 hours and cannot be ran manually. If a user wishes to run this manually, they can do so by running the following command:

```bash
import os
import requests

os.chdir('./uploads')

def main():
    # The URL of the endpoint with the list of assets
    url = "http://localhost:37238/assets"

    try:
        # Send a GET request to the URL
        response = requests.get(url)
        # Raise an exception if the request was not successful
        response.raise_for_status()

        # Parse the JSON response
        assets = response.json()

        # Validating the response to ensure it's a list of dictionaries
        if not isinstance(assets, list) or not all(isinstance(asset, dict) for asset in assets):
            print("Unexpected response format.")
            return

        # Extracting the file names from the JSON list
        allowed_file_names = {asset['file_name'] for asset in assets}

        # Get the list of files in the current directory
        current_files = os.listdir('.')

        # Loop through the files in the current directory
        for file in current_files:
            # Check if it's a file and not in the allowed list
            if os.path.isfile(file) and file not in allowed_file_names:
                print(f"Deleting file: {file}")
                os.remove(file)

        print("Operation completed.")

    except requests.exceptions.RequestException as e:
        print(f"HTTP Request failed: {e}")
    except Exception as e:
        print(f"An error occurred: {e}")

if __name__ == "__main__":
    main()

```
##### Asset
Whilst a user could delete directly from the filesystem, it is better to use the API to ensure the database is updated. A user could of course remove the row from the `assets` table.

```bash
curl -X DELETE http://localhost:37238/assets/{id}
```

To get the id, use the `GET` method or take note from the response when uploading.
#### Get (GET)
##### ID From Filename
It is possible to get the ID from the filename, this is useful for wikis that want to link to the file without using the asset_id. Using `id` makes sense for notes as they are all contained in the database, but it does not make sense for assets as they are stored in the filesystem anyway.

```bash
curl "http://localhost:37238/assets/id?filename=README.md"
```
```json
{"id":26}
```
##### List Files

```bash
curl http://localhost:37238/assets | jq
```

```json
[
  {
    "id": 2,
    "file_name": "asldkfj.org",
    "asset_type": "",
    "description": "This is a sample image upload",
    "created_at": "2024-10-26T07:14:01Z"
  },
  {
    "id": 3,
    "file_name": "asldkfj.org",
    "asset_type": "",
    "description": "This is a sample image upload",
    "created_at": "2024-10-26T07:14:01Z"
  }
]
```

##### Download
There is a dedicated endpoint for downloading files, there is not a server hosting all files. This is so the server can be extended for multiple users at some stage in the future.

```bash
curl -O -J "http://localhost:37238/assets/1/download"
cat asldkfj.org
```
```
This is a sample file
```

This link can be used to display the file in the browser (e.g. a Flask / Axum server):

```html
 <!-- For images -->
 <img src="http://localhost:37238/assets/1/download" alt="Description of image">

 <!-- For other file types -->
 <a href="http://localhost:37238/assets/2/download">Download File</a>

```

#### Clean Unused (Danger)
This checks for unused assets and removes them if they are not used.
#### Update
This is not implemented, but the client can delete and re-upload.
### Tags
#### Flat
##### Create

```sh
curl -X POST http://localhost:37238/tags \
  -H "Content-Type: application/json" \
  -d '{"name": "important"}' \
  -v -i
```

```json
{"id":7,"message":"Tag created successfully"}
```
##### Assign
To assign tag_id 3 to note_id 2:

```sh
curl -X POST http://localhost:37238/notes/2/tags \
      -H "Content-Type: application/json" \
      -d '{"tag_id": 3}'

```
##### Update
```sh
curl -X PUT -H "Content-Type: application/json" -d '{"name":"New Tag Name"}' http://localhost:37238/tags/1
```

```json
{"message":"Tag updated successfully"}
```
##### Delete

```sh
curl -X DELETE http://localhost:37238/tags/5
```

```
{"message":"Tag deleted successfully"}
```


##### Get
###### Tag and Notes
Create a list of notes

```sh
curl http://localhost:37238/tags/with-notes | jq
```

```json
[
  {
    "tag_id": 1,
    "tag_name": "important",
    "notes": null
  },
  {
    "tag_id": 3,
    "tag_name": "todo",
    "notes": [
      {
        "id": 2,
        "title": "Foo"
      }
    ]
  },
  {
    "tag_id": 2,
    "tag_name": "urgent",
    "notes": [
      {
        "id": 2,
        "title": "Foo"
      }
    ]
  },
  {
    "tag_id": 4,
    "tag_name": "done",
    "notes": null
  },
  {
    "tag_id": 5,
    "tag_name": "important",
    "notes": null
  }
]
```
###### Tag Names
It may be useful to only get the tag names (e.g. for a combo box):
```sh
curl http://localhost:37238/tags | jq
```

```json
[
  {
    "id": 4,
    "name": "done"
  },
  {
    "id": 1,
    "name": "important"
  },
  {
    "id": 5,
    "name": "important"
  },
  {
    "id": 3,
    "name": "todo"
  },
  {
    "id": 2,
    "name": "urgent"
  }
]
```
##### Search
This refers to searching for notes assigned to a tag.

This is left to the client, simply use a normal search on the notes endpoint and filter by tag. Less code to maintain and the client would need to get a list of tags first anyway.

For large corpuses it could be more efficient to postgres to do this, but for small corpuses it is not worth the overhead.

##### Filter
This is not implemented, but the client can use `fzf` to filter tags. It is not implemented because server latency will make it slow for palettes etc., particularly over, e.g. wireguard / tailscale.
#### Hierarchy
##### Create
```sh
curl -X POST http://localhost:37238/tags/hierarchy \
  -H "Content-Type: application/json" \
  -d '{"parent_tag_id": 10, "child_tag_id": 5}'

```
```json
{"message":"Tag hierarchy entry added successfully"}
```

##### Update
```sh
curl -X PUT http://localhost:37238/tags/hierarchy/5 \
      -H "Content-Type: application/json" \
      -d '{"parent_tag_id": 4}'
```

```json
{"message":"Tag hierarchy entry updated successfully"}
```
##### Delete
```sh
curl -X DELETE http://localhost:37238/tags/hierarchy/3
```
```json
{"message":"Tag hierarchy entry deleted successfully"}
```
##### Get (Tree)
To list the tags and the notes they contain:

```sh
curl http://localhost:37238/tags/tree
```

```json
[
  {
    "id": 1,
    "name": "important",
    "notes": null,
    "children": [
      {
        "id": 2,
        "name": "urgent",
        "notes": [
          {
            "id": 2,
            "title": "Foo"
          }
        ]
      },
      {
        "id": 3,
        "name": "todo",
        "notes": [
          {
            "id": 2,
            "title": "Foo"
          }
        ]
      }
    ]
  },
  {
    "id": 4,
    "name": "done",
    "notes": null
  },
  {
    "id": 5,
    "name": "important",
    "notes": null
  }
]

```
### Tasks
#### Task Entries
##### Create

```sh
 curl -X POST http://localhost:37238/tasks \
 -H "Content-Type: application/json" \
 -d '{
     "note_id": 1,
     "status": "todo",
     "effort_estimate": 2.5,
     "actual_effort": 0,
     "deadline": "2023-06-30T15:00:00Z",
     "priority": 3,
     "all_day": false,
     "goal_relationship": 4
 }'
```

```json
{"id":2,"message":"Task created successfully"}
```

##### Update

Updating the status, actual effort, and priority of a task:

```sh
 curl -X PUT http://localhost:37238/tasks/1 \
 -H "Content-Type: application/json" \
 -d '{
     "status": "done",
     "actual_effort": 3.5,
     "priority": 4
 }'
```

updating a single field:


```sh
 curl -X PUT http://localhost:37238/tasks/1 \
 -H "Content-Type: application/json" \
 -d '{"status": "wait"}'
```

```sh
 curl -X PUT http://localhost:37238/tasks/1 \
 -H "Content-Type: application/json" \
 -d '{
     "effort_estimate": 5.0,
     "deadline": "2023-07-15T14:00:00Z",
     "all_day": true
 }'
 ```

##### Delete

```bash
curl -X DELETE http://localhost:37238/tasks/1
```

```json
{"message":"Task deleted successfully"}
```

##### Get (tree)
```sh
curl http://localhost:37238/tasks/tree | jq
```

```
[
  {
    "id": 1,
    "title": "First note",
    "type": "",
    "children": [
      {
        "id": 2,
        "title": "Second note",
        "type": "block",
        "children": [
          {
            "id": 3,
            "title": "Third note",
            "type": "subpage"
          }
        ]
      }
    ]
  }
]
```

##### Get (flat)
The Get method returns tasks as a json like so:

```bash
curl http://localhost:37238/tasks/details | jq
```


```json
[
  {
    "id": 2,
    "note_id": 1,
    "status": "todo",
    "effort_estimate": 2.5,
    "actual_effort": 0,
    "deadline": "2023-06-30T15:00:00Z",
    "priority": 3,
    "all_day": false,
    "goal_relationship": 4,
    "created_at": "2024-10-20T10:31:58.269465Z",
    "modified_at": "2024-10-20T10:31:58.269465Z",
    "schedules": [
      {
        "id": 1,
        "start_datetime": "2023-06-01T09:00:00Z",
        "end_datetime": "2023-06-01T17:00:00Z"
      }
    ],
    "clocks": null
  },
  {
    "id": 7,
    "note_id": 4,
    "status": "todo",
    "effort_estimate": 2.5,
    "actual_effort": 0,
    "deadline": "2023-06-30T15:00:00Z",
    "priority": 3,
    "all_day": false,
    "goal_relationship": 4,
    "created_at": "2024-10-20T10:33:04.594991Z",
    "modified_at": "2024-10-20T10:33:04.594991Z",
    "schedules": null,
    "clocks": [
      {
        "id": 1,
        "clock_in": "2023-06-01T09:00:00Z",
        "clock_out": "2023-06-01T17:00:00Z"
      },
      {
        "id": 2,
        "clock_in": "2024-05-01T09:00:00Z",
        "clock_out": "2024-05-01T17:00:00Z"
      }
    ]
  }
]
```
#### Schedule
##### Create
```bash
 curl -X POST http://localhost:8080/task_schedules \
 -H "Content-Type: application/json" \
 -d '{
     "task_id": 2,
     "start_datetime": "2023-06-01T09:00:00Z",
     "end_datetime": "2023-06-01T17:00:00Z"
 }'
```

##### Update

```bash
curl -X PUT http://localhost:37238/task_schedules/1 \
 -H "Content-Type: application/json" \
 -d '{
     "start_datetime": "2022-06-02T10:00:00Z",
     "end_datetime": "2022-06-02T18:00:00Z"
 }'
```

```json
{"message":"Task schedule updated successfully"}
```

##### Delete
```bash
 curl -X DELETE http://localhost:37238/task_schedules/1
```
```json
{"message":"Task schedule deleted successfully"}
```
##### Get
This is handled by the task endpoint, as above.
#### Clock In
##### Create
```bash

curl -X POST http://localhost:37238/task_clocks \
 -H "Content-Type: application/json" \
 -d '{
     "task_id": 2,
     "clock_in": "2023-06-01T09:00:00Z",
     "clock_out": "2023-06-01T17:00:00Z"
 }'

```
```json
{"id":2,"message":"Task clock entry created successfully"}
```

##### Update
```bash
 curl -X PUT http://localhost:8080/task_clocks/1 -H "Content-Type: application/json" -d '{"clock_in": "2023-05-20T09:00:00Z",
 "clock_out": "2023-05-20T17:00:00Z"}'
```

```json
{"message":"Task clock entry updated successfully"}
```
##### Delete
```bash
curl -X DELETE http://localhost:8080/task_clocks/1
```

```json
{"message":"Task clock entry deleted successfully"}
```
##### Get
This is handled by the task endpoint, as above.
### Categories
Categories were abandoned in favor of tags. They are not implemented. There may be some leftover endpoints, these will be removed.
#### Get
```sh
curl http://localhost:37238/categories | jq
```
```json
[
  {
    "id": 3,
    "name": "Ideas"
  },
  {
    "id": 4,
    "name": "Journal"
  },
  {
    "id": 1,
    "name": "Personal"
  },
  {
    "id": 2,
    "name": "Work"
  }
]

```
#### Create

```sh
 curl -X POST http://localhost:37238/categories \
            -H "Content-Type: application/json" \
            -d '{"name": "New Category"}'

```
```json
{"id":5,"message":"Category created successfully"}
```
#### Assign

To assign category_id 3 to note_id 1:

```sh
curl -X POST http://localhost:37238/notes/1/categories \
      -H "Content-Type: application/json" \
      -d '{"category_id": 3}'


```
#### Assign hierarchy

```sh
curl -X POST http://localhost:37238/tags/hierarchy \
      -H "Content-Type: application/json" \
      -d '{"parent_tag_id": 1, "child_tag_id": 2}'
```

```json
{"id":4,"message":"Tag hierarchy entry added successfully"}
```

## Examples

### Task hierarchy

Consider the following list of tasks:

```sh
curl http://localhost:37238/tasks/details | jq
```

```json
[
  {
    "id": 2,
    "note_id": 1,
    "status": "todo",
    "effort_estimate": 2.5,
    "actual_effort": 0,
    "deadline": "2023-06-30T15:00:00Z",
    "priority": 3,
    "all_day": false,
    "goal_relationship": 4,
    "created_at": "2024-10-20T10:31:58.269465Z",
    "modified_at": "2024-10-20T10:31:58.269465Z",
    "schedules": null,
    "clocks": null
  },
  {
    "id": 7,
    "note_id": 4,
    "status": "todo",
    "effort_estimate": 2.5,
    "actual_effort": 0,
    "deadline": "2023-06-30T15:00:00Z",
    "priority": 3,
    "all_day": false,
    "goal_relationship": 4,
    "created_at": "2024-10-20T10:33:04.594991Z",
    "modified_at": "2024-10-20T10:33:04.594991Z",
    "schedules": null,
    "clocks": [
      {
        "id": 2,
        "clock_in": "2023-05-20T09:00:00Z",
        "clock_out": "2023-05-20T17:00:00Z"
      }
    ]
  }
]
```

If the note_hierarchy table looks like this:

```sql
select * from note_hierarchy;
```

```
 id | parent_note_id | child_note_id | hierarchy_type
----+----------------+---------------+----------------
 26 |              1 |             2 | subpage
 27 |              2 |             3 | subpage
 28 |              2 |             4 | subpage
(3 rows)
```

The hierarchy would appear like so:

```bash
curl http://localhost:37238/notes/tree | jq | x
```
```json
[
  {
    "id": 1,
    "title": "New Title",
    "type": "",
    "children": [
      {
        "id": 2,
        "title": "Foo",
        "type": "subpage",
        "children": [
          {
            "id": 3,
            "title": "Foo",
            "type": "subpage"
          },
          {
            "id": 4,
            "title": "New Note Title",
            "type": "subpage"
          }
        ]
      }
    ]
  }
]
```


Please be aware that `note_id: 2` is not a task itself, it will be included in the task tree to remain consistent with the `note_hierarchy` table, which remains valid for all trees.

```bash
curl http://localhost:37238/tasks/tree | jq
```

```json
[
  {
    "id": 1,
    "title": "New Title",
    "type": "",
    "children": [
      {
        "id": 2,
        "title": "Foo",
        "type": "subpage",
        "children": [
          {
            "id": 4,
            "title": "New Note Title",
            "type": "subpage"
          }
        ]
      }
    ]
  }
]
```

This could change in the future, where `note_id: 2` would be excluded from the output, and `id 4` would be directly linked as a child of `id 1`. Clients are encouraged to evaluate whether to maintain notes interspersed between tasks within the hierarchy.


Note also that the id here is the `note_id` rather than the `task_id`. This likely will change.
