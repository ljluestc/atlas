schema "public" {
}

table "users" {
  schema = schema.public
  column "id" {
    null = false
    type = int
  }
  primary_key {
    columns = [table.users.column.id]
  }
}

table "posts" {
  schema = schema.public
  column "id" {
    null = false
    type = int
  }
  column "user_id" {
    null = false
    type = int
  }
  column "title" {
    null = false
    type = varchar
  }
  foreign_key "owner_id" {
    columns = [table.posts.column.user_id]
    ref_columns = [table.users.column.id]
  }
  index "title_idx" {
    columns = [
      table.posts.column.title,
    ]
  }
}

enum "status" {
  schema = schema.public
  values = ["on", "off"]
}

table "meetings" {
  schema = schema.public
  column "id" {
    null = false
    type = int
  }
  column "room" {
    null = false
    type = int
  }
  column "during" {
    null = false
    type = "tsrange"
  }
  
  // EXCLUDE constraint that prevents overlapping meetings in the same room
  exclude "exclude_overlapping_meetings" {
    using = "gist"
    
    // Define the columns and their operators
    on {
      column = column.room
      operator = "="
    }
    
    on {
      column = column.during
      operator = "&&"  // The && operator tests for overlap
    }
  }
}

table "active_meetings" {
  schema = schema.public
  column "id" {
    null = false
    type = int
  }
  column "room" {
    null = false
    type = int
  }
  column "during" {
    null = false
    type = "tsrange"
  }
  column "active" {
    null = false
    type = bool
  }
  
  // EXCLUDE constraint with a WHERE predicate
  exclude "exclude_active_meetings" {
    using = "gist"
    
    on {
      column = column.room
      operator = "="
    }
    
    on {
      column = column.during
      operator = "&&"
    }
    
    where = "active = true"
  }
}

table "points" {
  schema = schema.public
  column "id" {
    null = false
    type = int
  }
  column "p" {
    null = false
    type = "point"
  }
  
  // EXCLUDE constraint with expressions
  exclude "exclude_points" {
    using = "gist"
    
    on {
      expr = "p <-> p"  // Distance between points
      operator = "<"
    }
    
    on {
      expr = "p"
      operator = "&&"
    }
  }
}
