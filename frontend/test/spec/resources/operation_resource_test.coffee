describe "OperationResource", ->
  describe "mostRecentRepositoryCheckouts", ->
    it "returns one checkout per repository",angular.mock.inject( (Operation) ->
      operation = new Operation({
        subject: {
          repositoryCheckouts: {
            refs: {
              "1ebdc6fa-a27e-4ab1-a763-4d9bc9e36ffd": [{
                Hash: "f1d2d2f924e986ac86fdf7b36c94bcdf32beec15",
                Ref: "refs/heads/master"
              }],
              "9b5aa008-56b7-4f17-95c4-ceca07755f6f": [{
                Hash: "e242ed3bffccdf271b7fbaf34ed72d089537b42f",
                Ref: "refs/heads/master"
              }]
            }
          }
        }
      })

      checkouts = operation.mostRecentRepositoryCheckouts
      expect(checkouts["1ebdc6fa-a27e-4ab1-a763-4d9bc9e36ffd"]).toBeDefined()
      expect(checkouts["9b5aa008-56b7-4f17-95c4-ceca07755f6f"]).toBeDefined()
    )

    it "returns the last checkout for each repository",angular.mock.inject( (Operation) ->
      repositoryUuid =
      lastRepositoryCheckout = {
        Hash: "e242ed3bffccdf271b7fbaf34ed72d089537b42f",
        Ref: "refs/heads/feature-branch"
      }
      operation = new Operation({
        subject: {
          repositoryCheckouts: {
            refs: {
              "abe18d5a-eff4-4092-b493-c235d4830912": [
                {
                  Hash: "f1d2d2f924e986ac86fdf7b36c94bcdf32beec15",
                  Ref: "refs/heads/master"
                },
                lastRepositoryCheckout
              ]
            }
          }
        }
      })

      checkouts = operation.mostRecentRepositoryCheckouts
      expect(checkouts["abe18d5a-eff4-4092-b493-c235d4830912"].Ref).toEqual(lastRepositoryCheckout.Ref)
      expect(checkouts["abe18d5a-eff4-4092-b493-c235d4830912"].Hash).toEqual(lastRepositoryCheckout.Hash)
    )

    it "returns checkout objects",angular.mock.inject( (Operation) ->
      repositoryUuid =
      operation = new Operation({
        subject: {
          repositoryCheckouts: {
            refs: {
              "abe18d5a-eff4-4092-b493-c235d4830912": [
                {
                  Hash: "f1d2d2f924e986ac86fdf7b36c94bcdf32beec15",
                  Ref: "refs/heads/master"
                }
              ]
            }
          }
        }
      })

      checkouts = operation.mostRecentRepositoryCheckouts
      expect(typeof checkouts["abe18d5a-eff4-4092-b493-c235d4830912"].shortHash).toBe("function")
      expect(typeof checkouts["abe18d5a-eff4-4092-b493-c235d4830912"].refName).toBe("function")
    )

  describe "RepositoryCheckout", () ->
    describe "shortHash", () ->
      it "returns the first seven characters of the hash",angular.mock.inject( (RepositoryCheckout) ->
        checkout = new RepositoryCheckout(
          Ref: 'refs/heads/master',
          Hash: '56252be02bc0c400d22aa6b60d1a578199c32211'
        )

        expect(checkout.shortHash()).toEqual('56252be')
      )

    describe "refName", () ->
      it "returns the last component of the ref",angular.mock.inject( (RepositoryCheckout)  ->
        checkout = new RepositoryCheckout(
          Ref: 'refs/heads/master',
          Hash: '56252be02bc0c400d22aa6b60d1a578199c32211'
        )

        expect(checkout.refName()).toEqual('master')
      )
