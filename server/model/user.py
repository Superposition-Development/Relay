from init import db, app
import time
from werkzeug.security import check_password_hash,generate_password_hash

class Users(db.Model):
    __tablename__ = "users"
    userID = db.Column(db.Text, unique=True)
    id = db.Column(db.Integer,primary_key=True)
    password = db.Column(db.Text)
    username = db.Column(db.Text)
    pfp = db.Column(db.Text)
    bio = db.Column(db.Text)
    date = db.Column(db.Text,default=time.time())
    
    def update(self, oldPW, newPW,username,pfp,bio):
        if not check_password_hash(self.password, oldPW):
            return "Password does not match"
        
        if not newPW.isspace() and newPW != "":
            self.password = generate_password_hash(newPW,method='pbkdf2:sha256')
            
        self.bio = bio
        
        if not username.isspace() and username != "":
            self.username = username
        
        if not pfp.isspace() and pfp != "":
            self.pfp = pfp
            
        db.session.commit()
        
        return "Success"
      
    
def initUserTable():
    with app.app_context():
        db.create_all()